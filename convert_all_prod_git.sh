#!/bin/bash
export NCPUS=10
export ES_URL="`cat ES_URL.prod.secret`"
#for idx in sds-gql-gql-git sds-hyperledger-aries-git sds-hyperledger-hyperledger-all-git sds-hyperledger-hyperledger-twgc-git sds-jdf-toip-git sds-tarscloud-git sds-yocto-git-for-merge sds-open-mainframe-project-zowe-git
for idx in sds-yocto-git-for-merge
do
  echo "converting ${idx}"
  ./es-convert-to-dads 'git' "${idx}" "${idx}-converted" || exit 1
  echo "converted ${idx}"
done
echo 'OK'
