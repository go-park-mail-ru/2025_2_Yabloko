@echo off
setlocal

REM Нагрузочное тестирование recommendation_service для ПОЛЬЗОВАТЕЛЯ С ИСТОРИЕЙ (profile-mode)
REM 1000 RPS, 60 секунд

if not exist results mkdir results

echo Running Vegeta recommend test for PROFILE user at 1000 RPS...

type ".\recommend_targets_profile.json" ^
  | vegeta attack -format=json -rate=1000 -duration=60s ^
    -connections=100 -max-connections=1000 -max-workers=1000 ^
    -output="results\opt\recommend_profile_1000_RPS.bin"

vegeta report -type=text "results\opt\recommend_profile_1000_RPS.bin" > "results\opt\recommend_profile_1000_RPS.txt"
vegeta plot "results\opt\recommend_profile_1000_RPS.bin" > "results\opt\recommend_profile_1000_RPS.html"

echo Done. Open results\opt\recommend_profile_1000_RPS.html in your browser.

endlocal
