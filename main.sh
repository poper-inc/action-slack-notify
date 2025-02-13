#!/usr/bin/env bash

export GITHUB_BRANCH=${GITHUB_REF##*heads/}
export SLACK_ICON=${SLACK_ICON:-"https://avatars0.githubusercontent.com/u/43742164"}
export SLACK_USERNAME=${SLACK_USERNAME:-"rtBot"}
export SLACK_AT_USERID=${SLACK_AT_USERID:-"userid"}
export CI_SCRIPT_OPTIONS="ci_script_options"
export SLACK_TITLE=${SLACK_TITLE:-"Message"}
export COMMIT_MESSAGE=$(cat "$GITHUB_EVENT_PATH" | jq -r '.head_commit.message')
export GITHUB_LAST_COMMIT_MESSAGE=$(git -C $GITHUB_WORKSPACE show-branch --no-name HEAD)
export GITHUB_LAST_COMMIT_AUTHOR=$(git -C $GITHUB_WORKSPACE log -1 --pretty=format:'%an')
export GITHUB_LAST_COMMIT_LONG_SHA=$(git -C $GITHUB_WORKSPACE rev-parse HEAD)
export GITHUB_LAST_COMMIT_SHORT_SHA=$(git -C $GITHUB_WORKSPACE rev-parse --short HEAD)
export GITHUB_REPO_NAME=${GITHUB_REPOSITORY:-"poper-inc"}
export TEST_DURATION=$(( $(date +%s) - $START_SECS ))"s"
export TIME_ZONE=${TIME_ZONE:-"UTC"}
export TEST_START=$(TZ=$TIME_ZONE date --date @$START_SECS +"%Y-%m-%d %H:%M:%S")

echo "$SLACK_MESSAGE" >> coverage.txt
if [[ $EXITCODE == "0" ]]; then
	export TEST_SUMMARY=$(grep "Summary:" -A3 coverage.txt)
else
	export TEST_SUMMARY=$(grep "ERRORS!" -A1 coverage.txt)" "$(grep "Summary" -A3 coverage.txt)
fi
rm -rf coverage.txt

hosts_file="$GITHUB_WORKSPACE/.github/hosts.yml"

if [[ -z "$SLACK_CHANNEL" ]]; then
	if [[ -f "$hosts_file" ]]; then
		user_slack_channel=$(cat "$hosts_file" | shyaml get-value "$CI_SCRIPT_OPTIONS.slack-channel" | tr '[:upper:]' '[:lower:]')
	fi
fi

if [[ -n "$user_slack_channel" ]]; then
	export SLACK_CHANNEL="$user_slack_channel"
fi

# Check vault only if SLACK_WEBHOOK is empty.
if [[ -z "$SLACK_WEBHOOK" ]]; then

	# Login to vault using GH Token
	if [[ -n "$VAULT_GITHUB_TOKEN" ]]; then
		unset VAULT_TOKEN
		vault login -method=github token="$VAULT_GITHUB_TOKEN" > /dev/null
	fi

	if [[ -n "$VAULT_GITHUB_TOKEN" ]] || [[ -n "$VAULT_TOKEN" ]]; then
		export SLACK_WEBHOOK=$(vault read -field=webhook secret/slack)
	fi
fi

if [[ -f "$hosts_file" ]]; then
	hostname=$(cat "$hosts_file" | shyaml get-value "$GITHUB_BRANCH.hostname")
	user=$(cat "$hosts_file" | shyaml get-value "$GITHUB_BRANCH.user")
	export HOST_NAME="\`$user@$hostname\`"
	export DEPLOY_PATH=$(cat "$hosts_file" | shyaml get-value "$GITHUB_BRANCH.deploy_path")

	temp_url=${DEPLOY_PATH%%/app*}
	export SITE_NAME="${temp_url##*sites/}"
    export HOST_TITLE="SSH Host"
fi

PR_SHA=$(cat $GITHUB_EVENT_PATH | jq -r .pull_request.head.sha)
[[ 'null' != $PR_SHA ]] && export GITHUB_SHA="$PR_SHA"

if [[ -n "$SITE_NAME" ]]; then
    export SITE_TITLE="Site"
fi


if [[ -z "$SLACK_MESSAGE" ]]; then
	export SLACK_MESSAGE="$COMMIT_MESSAGE"
fi

slack-notify "$@"
