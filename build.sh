#!/bin/sh

#abort on error
set -e

function usage
{
    echo "usage: build [--connectorOnly || -h]"
    echo "   ";
    echo "  -o | --connectorOnly : Only build connector";
    echo "  -s | --setup         : Setup the env";
    echo "  -h | --help          : This message";
}

function parse_args
{
  # positional args
  args=()

  # named args
  while [ "$1" != "" ]; do
      case "$1" in
          -o | --connectorOnly )        buildConn="$1";         shift;;
          -s | --setup )                setup="$1";             shift;;
          -h | --help )                 usage;                  exit;; # quit and show usage
          * )                           args+=("$1")             # if no match, add it to the positional args
      esac
  done

  # restore positional args
  set -- "${args[@]}"

}


function run
{
  parse_args "$@"
  if ! command -v node 2>/dev/null; then
    echo "node not found"
    exit
  fi
  if ! command -v yarn 2>/dev/null; then
    echo "yarn not found"
    exit
  fi
  if ! command -v go 2>/dev/null; then
    echo "go not found"
    exit
  fi
  if [ ! -z "$setup" ]; then
    echo "setting up"
    cd fefrontend && yarn
    exit
  fi
  if [ ! -z "$buildConn" ]; then
    echo "Build connector only"
    go build
    echo "Done building start with ./serial-port-json-server"
  else
    echo "build all"
    cd fefrontend && yarn build && cd ../ && go build
  fi
}



run "$@";