#!/bin/bash
export NCPUS=10
export ES_URL="`cat ES_URL.prod.secret`"
for idx in sds-hyperledger-aries-github-issue sds-hyperledger-hyperledger-twgc-github-issue sds-tarscloud-github-issue
do
  echo "converting ${idx}"
  ./es-convert-to-dads 'github/issue' "${idx}" "${idx}-converted" || exit 1
  echo "converted ${idx}"
done
echo 'OK'
