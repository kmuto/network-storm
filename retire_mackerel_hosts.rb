#!/usr/bin/env ruby
# Mackerel側の特定のサービス:ロールに存在するダミーホストを退役する
# 環境変数MACKEREL_APIKEY付きで呼び出す
#
# 第1引数: サービス名。省略するとswitches
# 第2引数: ロール名。省略するとtokyo
require 'json'

# 設定
SERVICE_NAME = ARGV[0] ? ARGV[0] : 'switches' # 対象のサービス名
ROLE_NAME = ARGV[1] ? ARGV[1] : 'tokyo' # 対象のロール名
BATCH_SIZE = 20     # 一度に退役させるホスト数
SLEEP_INTERVAL = 1  # 待機時間（秒）

unless ENV['MACKEREL_APIKEY']
  puts "Error: MACKEREL_APIKEY 環境変数付きで実行してください。"
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

# 2. IDの一覧を抽出
host_ids = hosts.map { |h| h['id'] }
total_count = host_ids.size

if total_count == 0
  puts "No hosts to retire."
  exit
end

puts "Found #{total_count} hosts. Starting retirement..."

# 3. 20個ずつに分割して実行
host_ids.each_slice(BATCH_SIZE).with_index do |slice, index|
  ids_to_retire = slice.join(' ')

  puts "[Batch #{(index + 1)}] Retiring: #{slice.first} ... #{slice.last} (#{slice.size} hosts)"

  # mkr retire ID1 ID2 ... を実行
  if ENV['DRY']
    puts("mkr retire --force #{ids_to_retire}")
  else
    system("mkr retire --force #{ids_to_retire}")
  end

  # 最後のバッチ以外はスリープを入れる
  if (index + 1) * BATCH_SIZE < total_count
    puts "Waiting for #{SLEEP_INTERVAL} second(s)..."
    sleep(SLEEP_INTERVAL)
  end
end

puts "All retirement tasks completed."
