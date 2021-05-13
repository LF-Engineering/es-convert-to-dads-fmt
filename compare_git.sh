#!/bin/bash
# ES1_URL="`cat ES_URL.prod.secret`"
# ES2_URL="`cat ES_URL.test.secret`"
# IDX1='bitergia-git_symphonyoss_200604_enriched_200930'
# IDX2='sds-cncf-k8s-git'
# KEY1=url_id
# KEY2=url_id
# ID1='finos/alloy/commit/f12d39aa02375258c444bd8815ba8bf621045615'
# ID2='kubernetes-csi/external-attacher/commit/28c782912bf7418f67601bccdda559cc7e64e880'
if [ -z "${ES1_URL}" ]
then
  export ES1_URL="`cat ES_URL.prod.secret`"
fi
if [ -z "${ES2_URL}" ]
then
  export ES2_URL="`cat ES_URL.test.secret`"
fi
if [ -z "${IDX1}" ] 
then
  export IDX1='bitergia-git_symphonyoss_200604_enriched_200930'
fi
if [ -z "${IDX2}" ] 
then
  export IDX2='sds-cncf-k8s-git'
fi
if [ -z "${KEY1}" ] 
then
  export KEY1='url_id'
fi
if [ -z "${KEY2}" ] 
then
  export KEY2='url_id'
fi
if [ -z "${ID1}" ] 
then
  export ID1='finos/alloy/commit/f12d39aa02375258c444bd8815ba8bf621045615'
fi
if [ -z "${ID2}" ] 
then
  export ID2='kubernetes-csi/external-attacher/commit/28c782912bf7418f67601bccdda559cc7e64e880'
fi
curl -s -H 'Content-Type: application/json' "${ES1_URL}/${IDX1}/_search" -d "{\"query\":{\"term\":{\"${KEY1}\":\"${ID1}\"}}}" | jq -rS '.' > p2o.json
curl -s -H 'Content-Type: application/json' "${ES2_URL}/${IDX2}/_search" -d "{\"query\":{\"term\":{\"${KEY2}\":\"${ID2}\"}}}" | jq -rS '.' > dads.json
cat p2o.json | sort -r | uniq > tmp && mv tmp p2o.txt
cat dads.json | sort -r | uniq > tmp && mv tmp dads.txt
echo "da-ds:" > report.txt
echo '-------------------------------------------' >> report.txt
cat dads.txt >> report.txt
echo '-------------------------------------------' >> report.txt
echo "p2o:" >> report.txt
echo '-------------------------------------------' >> report.txt
cat p2o.txt >> report.txt
echo '-------------------------------------------' >> report.txt
