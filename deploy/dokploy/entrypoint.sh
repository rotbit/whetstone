#!/bin/sh
set -eu

export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_STAT="${LOG_STAT:-false}"
export AUTH_ACCESS_EXPIRE="${AUTH_ACCESS_EXPIRE:-604800}"
export USER_RPC_ENDPOINT="${USER_RPC_ENDPOINT:-user-rpc:9001}"
export INTERVIEW_RPC_ENDPOINT="${INTERVIEW_RPC_ENDPOINT:-interview-rpc:9002}"
export QUESTION_RPC_ENDPOINT="${QUESTION_RPC_ENDPOINT:-question-rpc:9003}"

service="${WHETSTONE_SERVICE:?WHETSTONE_SERVICE is required}"
config_source="${CONFIG_SOURCE:-env}"
service_binary="${WHETSTONE_BINARY:-/usr/local/bin/whetstone-service}"
config_dir="${WHETSTONE_CONFIG_DIR:-/etc/whetstone}"

if [ "${config_source}" != "env" ] && [ "${config_source}" != "file" ]; then
  echo "CONFIG_SOURCE must be env or file" >&2
  exit 1
fi

case "${service}" in
  app-apis)
    config_file="${config_dir}/app-apis.yaml"
    if [ "${config_source}" = "env" ]; then
      : "${AUTH_ACCESS_SECRET:?AUTH_ACCESS_SECRET is required when CONFIG_SOURCE=env}"
      : "${WEBSOCKET_PUBLIC_URL:?WEBSOCKET_PUBLIC_URL is required when CONFIG_SOURCE=env}"
    fi
    ;;
  user-rpc)
    config_file="${config_dir}/user.yaml"
    ;;
  interview-rpc)
    config_file="${config_dir}/interview.yaml"
    ;;
  question-rpc)
    config_file="${config_dir}/question.yaml"
    ;;
  report-worker)
    exec "${service_binary}"
    ;;
  *)
    echo "unsupported WHETSTONE_SERVICE: ${service}" >&2
    exit 1
    ;;
esac

if [ ! -r "${config_file}" ]; then
  echo "configuration file is not readable: ${config_file}" >&2
  exit 1
fi

exec "${service_binary}" -f "${config_file}"
