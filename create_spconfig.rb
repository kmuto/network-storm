#!/usr/bin/env ruby
# pingのメトリック設定ファイルを生成する
# 環境変数MACKEREL_APIKEY付きで呼び出す
#
# 第1引数: IPアドレス一覧のCSVファイル名 (create_ip_and_delays.rbで作成)
# 第2引数: 書き出す設定ファイル名
require 'csv'

CSV_FILE = ARGV[0]
OUTPUT_FILE = ARGV[1]

unless ENV['MACKEREL_APIKEY']
  puts 'Error: MACKEREL_APIKEY 環境変数付きで実行してください。'
  exit
end

if CSV_FILE.nil? || !File.exist?(CSV_FILE)
  puts "Error: CSVファイル #{CSV_FILE} が見つかりません。"
  exit
end

unless OUTPUT_FILE
  puts 'Error: 書き出しファイル名が指定されていません。'
  exit
end

File.open(OUTPUT_FILE, 'w') do |f|
  f.puts "x-api-key: #{ENV['MACKEREL_APIKEY']}"
  f.puts 'collector:'

  CSV.foreach(CSV_FILE, headers: true) do |row|
    ip = row['ip']

    # テキストの書き出し
    f.puts "- host: #{ip}"
    f.puts "  custom-identifier: sw#{ip}"
    f.puts '  average: true'
  end
end

puts "#{OUTPUT_FILE} に設定を書き出しました。"
