#!/usr/bin/env ruby
# snmpstormの設定ファイルを生成する
#
# 第1引数: IPアドレス一覧のCSVファイル名 (create_ip_and_delays.rbで作成)
# 第2引数: 書き出す設定ファイル名
require 'csv'

PING_CSV = ARGV[0]
SNMP_CSV = ARGV[1]

if PING_CSV.nil? || !File.exist?(PING_CSV)
  puts "Error: #{PING_CSV} が見つかりません。"
  exit
end

total_ports = 0
host_count = 0

CSV.open(SNMP_CSV, "wb") do |csv|
  csv << ["ip", "if_count", "delay_ms"]

  CSV.foreach(PING_CSV, headers: true) do |row|
    ip = row['ip']
    
    # 1. ポート数 (最低1, 最大24)
    # 完全にランダムにするか、ある程度重みをつけるかは自由ですが、ここでは1..24で均等に振ります
    if_count = rand(1..24)

    # 2. ロス率による応答判定
    loss_rate = row['loss_rate'].to_f
    ping_median = row['median_ms'].to_f
    
    # loss_rateが0.5以上、または median_msが-1 (100%ロス) なら応答不能
    delay_ms = if loss_rate >= 0.5 || ping_median < 0
                 -1
               else
                 # 正常時はPingの遅延をそのまま採用（0未満にならないようmax(1)を入れる）
                 [ping_median.to_i, 1].max
               end

    csv << [ip, if_count, delay_ms]
    
    # 統計用
    total_ports += if_count
    host_count += 1
  end
end

puts "--- 変換完了 ---"
puts "読み込みホスト数: #{host_count} 台"
puts "生成ファイル　　: #{SNMP_CSV}"
puts "全ホスト合計ポート数: #{total_ports} ポート"
