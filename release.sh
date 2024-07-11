#!/bin/sh
if test -z "$1"; then
  echo Missing version number
  exit 1
fi
version="$1"
case $version in
    v*) 
        ;;
    *)
        version="v$version"
        ;;
esac
echo "Releasing $version"
if [ $(git tag -l "$version") ]; then
  while true; do
    read -p "Version already exists, are you sure? " yn
    case $yn in
        [Yy]* ) break;;
        [Nn]* ) exit;;
        * ) echo "Please answer yes or no.";;
    esac
  done
fi
hash=$(git rev-parse --short HEAD)
echo "Building for Linux..."
go build -ldflags "-X main.appVersion=$version -X main.gitHash=$hash" -o dist/doglog
echo "Building for Windows..."
GOOS=windows GOARCH=386 go build -ldflags "-X main.appVersion=$version -X main.gitHash=$hash" -o dist/doglog.exe
git tag -f -a "$1" -m "Release $version"
git push --force origin --tags
echo "Creating draft release..."
gh release create --draft --notes-from-tag "$version" "dist/doglog#doglog (Linux amd64 binary)" "dist/doglog.exe#doglog.exe (Windows amd64 binary)"
