#!/bin/sh

#abort on error
set -e

function usage
{
    echo "usage: build [--connectorOnly || -h]"
    echo "   ";
    echo "  -o | --connectorOnly : Only build connector";
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
  if [ ! -z "$buildConn" ]; then
    echo "Build connector only"
    git submodule update --remote && cd fefrontend && yarn && yarn build && cd ../ && go build && rm -rf release && mkdir release && cp ./serial-port-json-server.exe ./release && cp -r fefrontend/dist ./release/
    echo "Done building release"
  fi
}



run "$@";