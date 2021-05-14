#!/bin/bash
export NCPUS=10
export ES_URL="`cat ES_URL.prod.secret`"
for idx in sds-gql-gql-github-repository sds-hyperledger-aries-github-repository sds-hyperledger-hyperledger-twgc-github-repository sds-jdf-toip-github-repository sds-tarscloud-github-repository sds-open-mainframe-project-zowe-github-repository
do
  echo "converting ${idx}"
  ./es-convert-to-dads 'github/repository' "${idx}" "${idx}-converted" || exit 1
  echo "converted ${idx}"
done
echo 'OK'
