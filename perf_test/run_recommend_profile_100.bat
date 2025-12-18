@echo off
setlocal

REM Нагрузочное тестирование recommendation_service для ПОЛЬЗОВАТЕЛЯ С ИСТОРИЕЙ (profile-mode)
REM 100 RPS, 60 секунд

if not exist results mkdir results

echo Running Vegeta recommend test for PROFILE user at 100 RPS...

type ".\recommend_targets_profile.json" ^
  | vegeta attack -format=json -rate=100 -duration=60s ^
    -connections=10 -max-connections=100 -max-workers=100 ^
    -output="results\recommend_profile_100_RPS.bin"

vegeta report -type=text "results\recommend_profile_100_RPS.bin" > "results\recommend_profile_100_RPS.txt"
vegeta plot "results\recommend_profile_100_RPS.bin" > "results\recommend_profile_100_RPS.html"

echo Done. Open results\recommend_profile_100_RPS.html in your browser.

endlocal
