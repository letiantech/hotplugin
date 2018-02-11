#!/bin/sh
CURDIR=$(cd "$(dirname "$0")"; pwd)

if [ "$#" -lt 6 ]; then
    echo "Please provide a version number, a plugin name, a target plugin file name"
    exit 1
fi

#
# Get the command line args
#
while [[ $# > 1 ]]
do
    key="$1"

    case $key in
        -v|--version)
            VERSION="$2"
            shift
            ;;

        -t|--target)
            TARGET="$2"
            shift
            ;;
        -n|--name)
            NAME="$2"
            shift
            ;;
    esac

    shift
done

function usage()
{
    echo "usage:"
    echo "  ./newplugin.sh -v version -n name -t file"
}

if [[ $VERSION == "" ]]; then
    usage
elif [[ $TARGET == "" ]]; then
    usage
elif [[ $NAME == "" ]]; then
    usage
fi

cp $CURDIR/template.go $TARGET
sed -i "s/PLUGINNAME/$NAME/g" $TARGET
sed -i "s/PLUGINVERSION/$VERSION/g" $TARGET
