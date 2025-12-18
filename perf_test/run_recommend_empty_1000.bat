@echo off
setlocal

REM Нагрузочное тестирование recommendation_service для ПУСТОГО пользователя (random-mode)
REM 1000 RPS, 60 секунд

if not exist results mkdir results

echo Running Vegeta recommend test for EMPTY user at 1000 RPS...

type ".\recommend_targets_empty.json" ^
  | vegeta attack -format=json -rate=1000 -duration=60s ^
    -connections=100 -max-connections=1000 -max-workers=1000 ^
    -output="results\recommend_empty_1000_RPS.bin"

vegeta report -type=text "results\recommend_empty_1000_RPS.bin" > "results\recommend_empty_1000_RPS.txt"
vegeta plot "results\recommend_empty_1000_RPS.bin" > "results\recommend_empty_1000_RPS.html"

echo Done. Open results\recommend_empty_1000_RPS.html in your browser.

endlocal
