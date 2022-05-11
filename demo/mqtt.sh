#!/usr/bin/env bash

BROKER_HOST="${BROKER_HOST:-localhost}"
BROKER_PORT="${BROKER_PORT:-1883}"

ACTION="${1}"

mqtt_msg() {
    line="${1}"
    command="$(get_command "${line}")"
    args="$(get_args "${line}")"

	cat <<-EOF
	{"module":"gowon","msg":"${line}","nick":"tester","dest":"#gowon","command":"${command}","args":"${args}"}
	EOF
}

get_command() {
    line="${1}"

    if [[ "${line:0:1}" == "." ]]; then
        f="${line%% *}"
        echo "${f#\.}"
    else
        echo ""
    fi
}

get_args() {
    line="${1}"

    command="$(get_command "${line}")"

    if [[ "${line}" == ".${command}" ]]; then
        echo ""
    else
        echo "${line#\.${command}?}"
    fi

}

pub() {
    mqtt_msg "${@}" | mosquitto_pub -h "${BROKER_HOST}" -p "${BROKER_PORT}" -t "/gowon/input" -s
}

sub() {
    mosquitto_sub -h "${BROKER_HOST}" -p "${BROKER_PORT}" -t "/gowon/output" | jq -r '.msg'
}

case "${ACTION}" in
    pub)
        pub "${@:2}"
        ;;
    sub)
        sub "${@:2}"
        ;;
    *)
        echo "First argument must be either pub or sub" >&2
        ;;
esac
