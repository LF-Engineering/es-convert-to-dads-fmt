#!/bin/bash
tokens="`cat ../sync-data-sources/helm-charts/sds-helm/sds-helm/secrets/GITHUB_OAUTH.secret`"
tokens="${tokens//,/ }"
for f in $tokens
do
    f="$(echo "$f" | tr -d '\n' | tr -d '\r')"
    points="`curl -sH "Authorization: token $f" https://api.github.com/users/codertocat -I | grep --color=never x-ratelimit-limit`"
    remains="`curl -sH "Authorization: token $f" https://api.github.com/users/codertocat -I | grep --color=never x-ratelimit-remaining`"
    reset="`curl -sH "Authorization: token $f" https://api.github.com/users/codertocat -I | grep --color=never x-ratelimit-reset`"
    points="$(echo "$points" | tr -d '\n' | tr -d '\r')"
    remains="$(echo "$remains" | tr -d '\n' | tr -d '\r')"
    reset="$(echo "$reset" | tr -d '\n' | tr -d '\r')"
    echo "token $f $points $remains $reset"
done
