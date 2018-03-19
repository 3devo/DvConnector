#!/bin/sh

#abort on error
set -e
startDir=$(pwd)

echo -e " _____    ___     __   ___   ____   ____     ___     __  ______   ___   ____  "
echo -e "|     |  /  _]   /  ] /   \ |    \ |    \   /  _]   /  ]|      | /   \ |    \ "
echo -e "|   __| /  [_   /  / |     ||  _  ||  _  | /  [_   /  / |      ||     ||  D  )"
echo -e "|  |_  |    _] /  /  |  O  ||  |  ||  |  ||    _] /  /  |_|  |_||  O  ||    / "
echo -e "|   _] |   [_ /   \_ |     ||  |  ||  |  ||   [_ /   \_   |  |  |     ||    \ "
echo -e "|  |   |     |\     ||     ||  |  ||  |  ||     |\     |  |  |  |     ||  .  \\"
echo -e "|__|   |_____| \____| \___/ |__|__||__|__||_____| \____|  |__|   \___/ |__|\_|\t - $(git describe 2> /dev/null)"
echo -e ""
function usage
{
    echo "usage: build [--c || -f || -r 0.1 \"cool release\" ||  -h]"
    echo "   ";
    echo "  -c     | --connector     : Build connector";
    echo "  -f     | --frontend      : Build frontend";
    echo "  -r tag | --release tag   : Build release with tag";
    echo "  -h     | --help          : This help message";
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
    if ! command -v node > /dev/null; then
        echo "\e[31mnode not found">&2
        exit
    fi
    if ! command -v yarn > /dev/null; then
        echo "\e[31myarn not found">&2
        exit
    fi
    echo "* Updating submodule"
    if git submodule update --remote 2>/dev/null; then
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
    echo -e "\e[32mBUILDING CONNECTOR" >&2
    if ! command -v go > /dev/null; then
        echo -e "\e[31mgo not found" >&2
        exit
    fi
    if go build -o feconnector &> /dev/null; then
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
    # build_frontend

    echo "Building Linux amd64"
    cd $startDir
    mkdir release/feconnector-$1_linux_amd64
    cp -r fefrontend/dist ./release/feconnector-$1_linux_amd64 && cp -r default-files/* ./release/feconnector-$1_linux_amd64
    env GOOS=linux GOARCH=amd64 go build -tags="cli" -o release/feconnector-$1_linux_amd64/feconnector
    cd release
    tar -zcvf feconnector-$1_linux_amd64.tar.gz feconnector-$1_linux_amd64 1> /dev/null
    cd $startDir

    echo "Building Linux 386"
    mkdir release/feconnector-$1_linux_386
    cp -r fefrontend/dist ./release/feconnector-$1_linux_386 && cp -r default-files/* ./release/feconnector-$1_linux_386
    env GOOS=linux GOARCH=386 go build -tags="cli" -o release/feconnector-$1_linux_386/feconnector
    cd release
    tar -zcvf feconnector-$1_linux_386.tar.gz feconnector-$1_linux_386 1> /dev/null
    cd $startDir

    echo "Building Linux ARM (Raspi)"
    mkdir release/feconnector-$1_linux_arm
    cp -r fefrontend/dist ./release/feconnector-$1_linux_arm && cp -r default-files/* ./release/feconnector-$1_linux_arm
    env GOOS=linux GOARCH=arm go build -tags="cli" -o release/feconnector-$1_linux_arm/feconnector
    cd release
    tar -zcvf feconnector-$1_linux_arm.tar.gz feconnector-$1_linux_arm 1> /dev/null
    cd $startDir

    echo "Building Windows x32"
    mkdir release/feconnector-$1_windows_386
    cp -r fefrontend/dist ./release/feconnector-$1_windows_386 && cp -r default-files/* ./release/feconnector-$1_windows_386
    env GOOS=windows GOARCH=386 go build -v -o release/feconnector-$1_windows_386/feconnector.exe
    cd release/feconnector-$1_windows_386
    zip -r ../feconnector-$1_windows_386.zip * 1> /dev/null
    cd $startDir

    echo "Building Windows x64"
    mkdir release/feconnector-$1_windows_amd64
    cp -r fefrontend/dist ./release/feconnector-$1_windows_amd64 && cp -r default-files/* ./release/feconnector-$1_windows_amd64
    env GOOS=windows GOARCH=amd64 go build -v -o release/feconnector-$1_windows_amd64/feconnector.exe
    cd release/feconnector-$1_windows_amd64
    zip -r ../feconnector-$1_windows_amd64.zip * 1> /dev/null
    cd ../..

    echo "Building Darwin x64"
    mkdir release/feconnector-$1_darwin_amd64
    cp -r fefrontend/dist ./release/feconnector-$1_darwin_amd64 && cp -r default-files/* ./release/feconnector-$1_darwin_amd64
    env GOOS=darwin GOARCH=amd64 go build -tags="cli" -o release/feconnector-$1_darwin_amd64/feconnector
    cd release/feconnector-$1_darwin_amd64
    zip -r ../feconnector-$1_darwin_amd64.zip * &> /dev/null
    cd ../..

    github_release $1 "$*"
}

function github_release()
{
    echo "Creating release for feconnector $2"

    git tag -a v$1 -m "$2"
    git push origin v$1
    echo ""
    echo "Before creating release"
    github-release info

    github-release release \
    --tag v$1 \
    --name "Feconnector" \
    --description "A server that can connect to the next 1.0 for control and logging" \

    echo ""
    echo "After creating release"
    github-release info

    echo ""
    echo "Uploading binaries"

    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_linux_amd64.tar.gz" \
    --file release/feconnector-$1_linux_amd64.tar.gz
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_linux_386.tar.gz" \
    --file release/feconnector-$1_linux_386.tar.gz
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_linux_arm.tar.gz" \
    --file release/feconnector-$1_linux_arm.tar.gz
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_windows_386.zip" \
    --file release/feconnector-$1_windows_386.zip
    github-release upload \
    --tag v$1 \
    --name "feconnector-$1_windows_amd64.zip" \
    --file release/feconnector-$1_windows_amd64.zip
    echo ""
    echo "Done"
    echo "Release can be found at -> https://github.com/3devo/FeConnector/releases"
    exit
}


function run
{
    if  [ "$#" -lt 1 ]; then
        usage
        exit
    fi
    parse_args "$@"
}

run "$@";