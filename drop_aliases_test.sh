#!/bin/bash
export ES_URL="`cat ES_URL.test.secret`"
for idx in bitergia-github_cloudfoundry_180322_enriched_191002 bitergia-github_symphonyoss_180322_enriched_200930
do
  curl -s -XDELETE "${ES_URL}/${idx}/_alias/sds-*-github-issue"
done
for idx in bitergia-git_cloudfoundry_181221_enriched_191007 bitergia-git_onap_191112_enriched_191112 bitergia-git_opendaylight_181221_enriched_200521 bitergia-git_opnfv_181220_enriched_191007 bitergia-git_symphonyoss_200604_enriched_200930 bitergia-git_yoctoproject_200109_enriched_200109
do
  curl -s -XDELETE "${ES_URL}/${idx}/_alias/sds-*-git"
done
