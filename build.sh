#!/bin/sh

#abort on error
set -e
startDir=$(pwd)
echo "" > $startDir/error.log
BUILDTOOLS=(node yarn go github-release zip tar)
echo -e " _____    ___     __   ___   ____   ____     ___     __  ______   ___   ____  "
echo -e "|     |  /  _]   /  ] /   \ |    \ |    \   /  _]   /  ]|      | /   \ |    \ "
echo -e "|   __| /  [_   /  / |     ||  _  ||  _  | /  [_   /  / |      ||     ||  D  )"
echo -e "|  |_  |    _] /  /  |  O  ||  |  ||  |  ||    _] /  /  |_|  |_||  O  ||    / "
echo -e "|   _] |   [_ /   \_ |     ||  |  ||  |  ||   [_ /   \_   |  |  |     ||    \ "
echo -e "|  |   |     |\     ||     ||  |  ||  |  ||     |\     |  |  |  |     ||  .  \\"
echo -e "|__|   |_____| \____| \___/ |__|__||__|__||_____| \____|  |__|   \___/ |__|\_| - $(git describe 2>> $startDir/error.log)"
echo -e ""

error() {
    local parent_lineno="$1"
    local message="$2"
    local code="${3:-1}"
    echo -e "\e[31mError on or near line ${parent_lineno}"
    echo "Please check the error.log for more information"
    exit
}

trap 'error ${LINENO}' ERR

function usage
{
    echo "usage: ./build.sh [-c || -f || -r 0.1 \"cool release\" ||  -h]"
    echo "   ";
    echo "  -c             | --connector           : Build connector";
    echo "  -f             | --frontend            : Build frontend";
    echo "  -r tag message | --release tag message : Build release with tag and message";
    echo "  -h             | --help                : This help message";
}

function parse_args
{
    # positional args
    args=()
    # named args
    while [ "$1" != "" ]; do
        case "$1" in
            -c | --connector )            build_connector;        shift;;
            -f | --frontend )             build_frontend;         shift;;
            -r | --release )              release $2 $3;          shift;;
            -h | --help )                 usage;                  exit;; # quit and show usage
        esac
    done

    # restore positional args
    set -- "${args[@]}"
}

function build_frontend
{
    cd $startDir
    echo -e "\e[32mBUILDING FRONTEND"
    echo "* Updating submodule"
    if git submodule update --remote 1>/dev/null 2>>$startDir/error.log; then
        echo "  - Update complete"
    else
        echo -e " \e[31m - Failed updating submodule"  >&2
        exit
    fi

    echo "* Building frontend"
    if cd fefrontend && yarn  && yarn build; then
        echo "  - Done building frontend"
    else
        echo -e "\e[31m - Failed building frontend"  >&2
        exit
    fi
}


function build_connector
{
    cd $startDir
    echo -e "\e[32mBUILDING CONNECTOR\e[31m" >&2
    if ! command -v go > /dev/null; then
        echo -e "\e[31mgo not found" >&2
        exit
    fi
    if go build -o feconnector.exe 1>/dev/null 2>> $startDir/error.log; then
        echo "  - Done building connector"
    else
        echo -e "\e[31m - Failed building connector"  >&2
        exit
    fi
}

function release()
{
    if [ "$#" -lt 2 ]; then
        echo "You need to pass in the version number and message ./build.sh -r 0.1 cool release"
        exit
    fi
    rm -rf release && mkdir -p release
    echo -e "\e[32mBUILDING RELEASE - \e[33m$*"
    echo -e "\e[33m-------------------------\e[32m"
    build_frontend

    if ! build_linux $1 "amd64"; then
        echo -e "\e[31m  LINUX AMD64 BUILD FAILED"
    fi
    if ! build_linux $1 "386"; then
        echo -e "\e[31m  LINUX 386 BUILD FAILED"
    fi
    if ! build_linux $1 "arm"; then
        echo -e "\e[31m  LINUX ARM BUILD FAILED"
    fi

    if ! build_windows $1 "386"; then
        echo -e "\e[31m  WINDOWS x32 BUILD FAILED"
    fi

    if ! build_windows $1 "amd64"; then
        echo -e "\e[31m  WINDOWS x64 BUILD FAILED"
    fi

    if ! build_osx $1; then
        echo -e "\e[31m  Darwin BUILD FAILED"
    fi

    github_release $1 "$*" || error $LINENO
}

function github_release()
{
    echo "Creating release for feconnector $2"

    git tag -a v$1 -m "$2" 1>/dev/null 2>> $startDir/error.log || error $LINENO
    git push origin v$1 1>/dev/null 2>> $startDir/error.log || error $LINENO
    echo ""
    echo "Before creating release"
    github-release info || error $LINENO

    github-release release \
    --tag v$1 \
    --name "Feconnector" \
    --description "A server that can connect to the next 1.0 for control and logging" \ || error $LINENO

    echo ""
    echo "After creating release"
    github-release info || error $LINENO

    echo ""
    echo "Uploading binaries"

    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_linux_amd64.tar.gz" \
    --file release/feconnector-$1_linux_amd64.tar.gz || error $LINENO
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_linux_386.tar.gz" \
    --file release/feconnector-$1_linux_386.tar.gz || error $LINENO
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_linux_arm.tar.gz" \
    --file release/feconnector-$1_linux_arm.tar.gz || error $LINENO
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_windows_386.zip" \
    --file release/feconnector-$1_windows_386.zip || error $LINENO
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_windows_amd64.zip" \
    --file release/feconnector-$1_windows_amd64.zip || error $LINENO
    echo ""
    echo "Done"
    echo "Release can be found at -> https://github.com/3devo/FeConnector/releases"
    exit
}


function build_linux()
{
    echo -e "\e[32mBuilding Linux $2"
    cd $startDir
    mkdir release/feconnector-$1_linux_$2
    cp -r fefrontend/dist ./release/feconnector-$1_linux_$2 1>/dev/null 2>> $startDir/error.log && cp -r default-files/* ./release/feconnector-$1_linux_$2 || error $LINENO
    env GOOS=linux GOARCH=$2 go build -tags="cli" -o release/feconnector-$1_linux_$2/feconnector 1>/dev/null 2>> $startDir/error.log  || error $LINENO
    cd release
    tar -zcvf feconnector-$1_linux_$2.tar.gz feconnector-$1_linux_$2 1>/dev/null 2>> $startDir/error.log || error $LINENO
    cd $startDir
}

function build_windows()
{
    echo -e "\e[32mBuilding Windows $2"
    mkdir release/feconnector-$1_windows_$2
    cp -r fefrontend/dist ./release/feconnector-$1_windows_$2 1>/dev/null 2>> $startDir/error.log  && cp -r default-files/* ./release/feconnector-$1_windows_$2 1>/dev/null 2>> $startDir/error.log || error $LINENO
    env GOOS=windows GOARCH=$2 go build -v -o release/feconnector-$1_windows_$2/feconnector.exe 1>/dev/null 2>> $startDir/error.log || error $LINENO
    cd release/feconnector-$1_windows_$2
    zip -r ../feconnector-$1_windows_$2.zip * 1>/dev/null 2>> $startDir/error.log || error $LINENO
    cd ../..
}

function build_osx()
{
    echo -e "\e[32mBuilding Darwin x64"
    mkdir release/feconnector-$1_darwin_amd64
    cp -r fefrontend/dist ./release/feconnector-$1_darwin_amd64 1>/dev/null 2>> $startDir/error.log && cp -r default-files/* ./release/feconnector-$1_darwin_amd64 1>/dev/null 2>> $startDir/error.log || error $LINENO
    env GOOS=darwin GOARCH=amd64 go build -tags="cli" -o release/feconnector-$1_darwin_amd64/feconnector 1>/dev/null 2>> $startDir/error.log || error $LINENO
    cd release/feconnector-$1_darwin_amd64
    zip -r ../feconnector-$1_darwin_amd64.zip * 1>/dev/null 2>> $startDir/error.log || error $LINENO
    cd ../..
}

function run
{
    error=0
    for i in ${BUILDTOOLS[@]}; do
        if ! command -v ${i} > /dev/null; then
            echo -e "\e[31mCommand \"${i}\" is missing in your path">&2
            local error=1
        fi
    done
    echo ""
    if [ $error -gt 0 ]; then
        echo -e "\e[31mError occured some binaries are missing in your path"
        exit
    fi

    if  [ "$#" -lt 1 ]; then
        usage
        exit
    fi
    parse_args "$@"
}

run "$@";
