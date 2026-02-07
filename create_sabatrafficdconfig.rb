#!/usr/bin/env ruby
# sabatrafficdの設定ファイルを生成する
# 環境変数MACKEREL_APIKEY付きで呼び出す
#
# 第1引数: 書き出す設定ファイル名
# 第2引数: サービス名。省略するとswitches
# 第3引数: ロール名。省略するとtokyo
require 'json'
require 'ipaddr'

# 設定
CONFIG_FILE = ARGV[0]
SERVICE_NAME = ARGV[1] ? ARGV[1] : 'switches' # 対象のサービス名
ROLE_NAME = ARGV[2] ? ARGV[2] : 'tokyo' # 対象のロール名

unless ENV['MACKEREL_APIKEY']
  puts 'Error: MACKEREL_APIKEY 環境変数付きで実行してください。'
  exit
end

# 1. mkr hosts を実行してJSONを取得
puts "Fetching hosts for role: #{SERVICE_NAME}:#{ROLE_NAME}..."
json_data = `mkr hosts --service #{SERVICE_NAME} --role #{ROLE_NAME}`

if json_data.empty?
  puts "No hosts found or error executing mkr command."
  exit
end

begin
  hosts = JSON.parse(json_data)
rescue JSON::ParserError => e
  puts "JSON parse error: #{e.message}"
  exit
end

sorted_hosts = hosts.sort_by do |host|
  ip_str = host['name'].sub(/^sw/, '')
  begin
    IPAddr.new(ip_str).to_i
  rescue
    0 # IP形式でない場合は先頭へ
  end
end

File.open(CONFIG_FILE, 'wb') do |f|
  f.puts "x-api-key: #{ENV['MACKEREL_APIKEY']}"
  f.puts 'collector:'

  sorted_hosts.each do |host|
    host_id = host['id']
    name = host['name']

    # name (例: sw10.0.0.200) から "sw" を除いたIP部分を抽出
    # 文字列の先頭から "sw" を削除する
    ip_address = name.sub(/^sw/, '')

    # YAML形式で出力
    f.puts "- host-id: #{host_id}"
    f.puts "  hostname: sw#{ip_address}"
    f.puts "  host: #{ip_address}"
    f.puts '  port: 1611'
    f.puts '  community: public'
    f.puts '  mibs:'
    f.puts '    - ifHCInOctets'
    f.puts '    - ifHCOutOctets'
    f.puts '    - ifInDiscards'
    f.puts '    - ifOutDiscards'
    f.puts '    - ifInErrors'
    f.puts '    - ifOutErrors'
  end
end
