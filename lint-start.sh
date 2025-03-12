#!/usr/bin/env bash
# 스크립트 내에서 명령어가 실패하면 즉시 종료
set -e

# 제대로 설치되어 있다면 문제는 없음.
# 나중에 주석으로 코드에 주석으로 추가하는 방향으로 가는 것은 어떤지 생각해보자.
#golangci-lint run --out-format json > lint_results.json
golangci-lint run