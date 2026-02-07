#!/usr/bin/env ruby
# sim_ping.pyおよびほかのツール向けのIPアドレス一覧を生成する
# sim_ping.py用のランダムな状態も入っている
#
# 第1引数: 生成するCSVファイル名
# 第2引数: IPアドレス数。省略すると1000
# 第3引数: タイムアウトのホストの頻度。省略すると0.05 (5%)
# 第4引数: 何番目から開始するか
require 'csv'

FILE_NAME = ARGV[0]
TOTAL_HOSTS = ARGV[1] ? ARGV[1].to_i : 1000
TIMEOUT_CHANCE = ARGV[2] ? ARGV[2].to_f : 0.05
START_INDEX = ARGV[3] ? ARGV[3].to_i : 0

unless FILE_NAME
  puts "Error: 書き出すCSVファイル名を指定してください。"
  exit
end

CSV.open(FILE_NAME, "wb") do |csv|
  csv << ["ip", "median_ms", "jitter_ms"]

  (START_INDEX...(START_INDEX + TOTAL_HOSTS)).each do |i|
    # 通算インデックス i から IPアドレスを計算
    # 1つのオクテットにつき254個 (1..254) 使える計算
    octet3 = i / 254
    octet4 = (i % 254) + 1

    # 10.0.255.x を超える場合のガード（必要であれば）
    if octet3 > 254
      puts "Warning: IPアドレスの範囲(10.0.254.254)を超えました。中断します。"
      break
    end

    ip = "10.0.#{octet3}.#{octet4}"

    # 数値生成ロジック
    if rand < TIMEOUT_CHANCE
      median, jitter = -1, 0
    else
      case rand
      when 0..0.8   then median, jitter = rand(5..50), rand(1..10)
      when 0.8..0.95 then median, jitter = rand(100..500), rand(20..100)
      else               median, jitter = rand(1000..3000), rand(200..800)
      end
    end

    csv << [ip, median, jitter]
  end
end

puts "#{FILE_NAME} に インデックス #{START_INDEX} から #{TOTAL_HOSTS} 件のデータを生成しました。"
