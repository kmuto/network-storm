#!/usr/bin/env ruby
# Mackerel側のダミーホストを生成する
# 環境変数MACKEREL_APIKEY付きで呼び出す
#
# 第1引数: IPアドレス一覧のCSVファイル名 (create_ip_and_delays.rbで作成)
# 第2引数: サービス:ロール の文字列。省略すると switches:tokyo
require 'csv'

# 設定
CSV_FILE = ARGV[0]
SERVICE_ROLE = ARGV[1] ? ARGV[1] : 'switches:tokyo'
WAIT_TIME = 0.5 # 0.5秒待機

unless ENV['MACKEREL_APIKEY']
  puts 'Error: MACKEREL_APIKEY 環境変数付きで実行してください。'
  exit
end

if CSV_FILE.nil? || !File.exist?(CSV_FILE)
  puts "Error: CSVファイル #{CSV_FILE} が見つかりません。"
  exit
end

# CSVからIPアドレスを読み込み
ip_list = []
CSV.foreach(CSV_FILE, headers: true) do |row|
  ip_list << row['ip']
end

total_count = ip_list.size
puts "#{total_count} 件のホストを順次作成します（間隔: #{WAIT_TIME}秒）..."

ip_list.each_with_index do |ip, index|
  identifier = "sw#{ip}"

  # 1件ずつ作成
  success = true
  if ENV['DRY']
    puts("mkr create -R #{SERVICE_ROLE} --customIdentifier #{identifier} #{identifier}")
  else
    success = system("mkr create -R #{SERVICE_ROLE} --customIdentifier #{identifier} #{identifier}")
  end

  if success
    puts "[#{index + 1}/#{total_count}] Created: #{identifier}"
  else
    puts "[#{index + 1}/#{total_count}] Failed: #{identifier}"
  end

  # 最後のホスト以外は待機
  sleep(WAIT_TIME) if index + 1 < total_count
end

puts 'すべての作成処理が終了しました。'
