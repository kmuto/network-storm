#!/usr/bin/env ruby
# sabatrafficdの設定ファイルを生成する（ダミーバージョン）
# 環境変数MACKEREL_APIKEY付きで呼び出す
#
# 第1引数: IPアドレス一覧のCSVファイル名 (create_ip_and_delays.rbで作成)
# 第2引数: 書き出す設定ファイル名
# 第3引数: サービス名。省略するとswitches
# 第4引数: ロール名。省略するとtokyo
require 'json'
require 'ipaddr'
require 'csv'

# 設定
CSV_FILE = ARGV[0]
CONFIG_FILE = ARGV[1]
SERVICE_NAME = ARGV[2] ? ARGV[2] : 'switches' # 対象のサービス名
ROLE_NAME = ARGV[3] ? ARGV[3] : 'tokyo' # 対象のロール名

unless ENV['MACKEREL_APIKEY']
  puts 'Error: MACKEREL_APIKEY 環境変数付きで実行してください。'
  exit
end

if CSV_FILE.nil? || !File.exist?(CSV_FILE)
  puts "Error: CSVファイル #{CSV_FILE} が見つかりません。"
  exit
end

sorted_hosts = []
c = 1
CSV.foreach(CSV_FILE, headers: true) do |row|
  ip = row['ip']
  sorted_hosts << { "id" => sprintf('s%010d', c), "name" => "sw#{ip}" }
  c += 1
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
