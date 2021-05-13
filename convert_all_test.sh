#!/bin/bash
export ES_URL="`cat ES_URL.test.secret`"
if [ ! -z "$DROP" ]
then
  curl -s -XDELETE "${ES_URL}/bitergia-*-git*-converted"
  echo ''
  echo "Dropped all converted indices"
fi
if [ -z "$NCPUS" ]
then
  export NCPUS=16
fi

# GitHub issues
echo 'github/issue'
for idx in bitergia-github_cloudfoundry_180322_enriched_191002 bitergia-github_symphonyoss_180322_enriched_200930
do
  echo "converting $idx"
  ./es-convert-to-dads 'github/issue' "$idx" "${idx}-converted" || exit 1
  echo "converted $idx"
done

# Git
echo 'git'
for idx in bitergia-git_cloudfoundry_181221_enriched_191007 bitergia-git_onap_191112_enriched_191112 bitergia-git_opendaylight_181221_enriched_200521 bitergia-git_opnfv_181220_enriched_191007 bitergia-git_symphonyoss_200604_enriched_200930 bitergia-git_yoctoproject_200109_enriched_200109
do
  echo "converting $idx"
  ./es-convert-to-dads git "$idx" "${idx}-converted" || exit 3
  echo "converted $idx"
done
