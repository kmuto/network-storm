#!/usr/bin/env ruby
# sim_ping.pyおよびほかのツール向けのIPアドレス一覧を生成する
# sim_ping.py用のランダムな状態も入っている
#
# 第1引数: 生成するCSVファイル名
# 第2引数: IPアドレス数。省略すると1000
# 第3引数: 先頭IPアドレス2オクテット。省略すると10.0
# 第4引数: 何番目から開始するか
require 'csv'

FILE_NAME = ARGV[0]
TOTAL_HOSTS = ARGV[1] ? ARGV[1].to_i : 1000
START_IP = ARGV[2] ? ARGV[2] : '10.0'
START_INDEX = ARGV[3] ? ARGV[3].to_i : 0

unless FILE_NAME
  puts "Error: 書き出すCSVファイル名を指定してください。"
  exit
end

CSV.open(FILE_NAME, "wb") do |csv|
  csv << ["ip", "median_ms", "jitter_ms", "loss_rate"]

  (START_INDEX...(START_INDEX + TOTAL_HOSTS)).each do |i|
    # 通算インデックス i から IPアドレスを計算
    # 1つのオクテットにつき254個 (1..254) 使える計算
    octet3 = i / 254
    octet4 = (i % 254) + 1

    # 10.0.255.x を超える場合のガード（必要であれば）
    if octet3 > 254
      puts "Warning: IPアドレスの範囲(#{START_IP}.254.254)を超えました。中断します。"
      break
    end

    ip = "#{START_IP}.#{octet3}.#{octet4}"

    # ネットワーク状態のシミュレーション
    case rand
    when 0..0.8 # 正常: 低遅延・ロスなし
      median, jitter, loss = rand(5..30), rand(1..5), 0.0
    when 0.8..0.9 # 不安定: 中遅延・時々ロス
      median, jitter, loss = rand(100..300), rand(50..150), rand(0.01..0.05).round(3)
    when 0.9..0.97 # 輻輳: 高遅延・高いロス
      median, jitter, loss = rand(500..1500), rand(200..500), rand(0.1..0.3).round(3)
    else # 障害: ほぼ届かない
      median, jitter, loss = rand(2000..3000), 500, rand(0.8..1.0).round(3)
    end

    if ENV['IDEAL']
      median = 0
      jitter = 0
      loss = 0
    end

    csv << [ip, median, jitter, loss]
  end
end

puts "#{FILE_NAME} に インデックス #{START_INDEX} から #{TOTAL_HOSTS} 件のデータを生成しました。"
