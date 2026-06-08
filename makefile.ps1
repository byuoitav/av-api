$COMMAND = $args[0]
$ErrorActionPreference = "Stop"

$NAME = "av-api"
$OWNER = "byuoitav"
$PKG = "github.com/$OWNER/$NAME"
$DOCKER_URL = "ghcr.io"
$DOCKER_PKG = "$DOCKER_URL/$OWNER/$NAME"

Write-Output "PKG: $PKG"
Write-Output "DOCKER_PKG: $DOCKER_PKG"

$PRD_TAG_REGEX = "v[0-9]+\.[0-9]+\.[0-9]+"
$DEV_TAG_REGEX = "v[0-9]+\.[0-9]+\.[0-9]+-.+"

$COMMIT_HASH = Invoke-Expression "git rev-parse --short HEAD"
$TAG = Invoke-Expression "git rev-parse --short HEAD"
try {
    $NEW_TAG = Invoke-Expression "git describe --exact-match --tags HEAD"
    Write-Output "NEW_TAG: $NEW_TAG.Length"
    if ($NEW_TAG.Length -gt 0) {
        $TAG = $NEW_TAG
        Write-Output "The repo contains a tag: $TAG"
    }
}
catch {
    Write-Output "The repo does not contain a tag"
}

Write-Output "The TAG is: $TAG"

function Invoke-Checked {
    param (
        [string]$Command
    )

    Invoke-Expression $Command
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: $Command"
    }
}

function Get-PkgList {
    $PKG_LIST = Invoke-Expression "go list $PKG/..."
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: go list $PKG/..."
    }

    Write-Output "PKG_LIST: $PKG_LIST"
    return $PKG_LIST
}

function All {
    Write-Output "All"
}

function Test {
    Write-Output "Test"
    $PKG_LIST = Get-PkgList
    Invoke-Checked "go test -v -race $PKG_LIST"
}

function Test-cov {
    Write-Output "Test-cov"
    $PKG_LIST = Get-PkgList
    Invoke-Checked "go test -coverprofile=coverage.txt -covermode=atomic $PKG_LIST"
}

function Lint {
    Write-Output "Lint"
    Invoke-Checked "golangci-lint run --test=false"
}

function Deps {
    Write-Output "Downloading Dependencies"
    Invoke-Checked "go mod download"
}

function Build {
    Write-Output "Build"

    New-Item -Path dist -ItemType Directory -Force | Out-Null
    Copy-Item "version.txt" -Destination "dist\version.txt" -Force

    Write-Output "*****************************************"
    Write-Output "Building for linux-arm"
    try {
        Set-Item -Path env:CGO_ENABLED -Value 0
        Set-Item -Path env:GOOS -Value "linux"
        Set-Item -Path env:GOARCH -Value "arm"
        Set-Item -Path env:GOARM -Value "7"
        Invoke-Checked "go build -v -o dist/${NAME}-arm"

        if (-not (Test-Path -Path "dist\${NAME}-arm")) {
            throw "Build completed without creating dist\${NAME}-arm"
        }
    }
    finally {
        Set-Item -Path env:GOOS -Value "windows"
        Set-Item -Path env:GOARCH -Value "amd64"
        Remove-Item env:GOARM -ErrorAction SilentlyContinue
    }

    Write-Output "Build output is located in ./dist/."
}

function Cleanup {
    Write-Output "Clean"
    Invoke-Checked "go clean"

    if (Test-Path -Path "dist") {
        Remove-Item dist -Recurse
        Write-Output "Recursively deleted dist/"
    }
    else {
        Write-Output "No dist directory to delete"
    }

    if (Test-Path -Path "${NAME}-bin") {
        Remove-Item "${NAME}-bin"
        Write-Output "Deleted legacy ${NAME}-bin"
    }
    if (Test-Path -Path "${NAME}-arm") {
        Remove-Item "${NAME}-arm"
        Write-Output "Deleted legacy ${NAME}-arm"
    }
}

function DockerFunc {
    # can not just be docker because it creates an infinite loop
    Write-Output "Function Docker      Commit Hash: $COMMIT_HASH     Tag: $TAG"

    if (-not (Test-Path -Path "dist\${NAME}-arm")) {
        throw "Missing dist\${NAME}-arm. Run Build before Docker."
    }

    if ($COMMIT_HASH -eq $TAG) {
        Write-Output "Building dev containers with tag $COMMIT_HASH"

        Write-Output "Building container $DOCKER_PKG/${NAME}-arm-dev:$COMMIT_HASH"
        Invoke-Checked "docker build -f dockerfile-arm --platform linux/arm/v7 --build-arg NAME=${NAME}-arm -t $DOCKER_PKG/${NAME}-arm-dev:$COMMIT_HASH dist"
    }
    elseif ($TAG -match $DEV_TAG_REGEX) {
        Write-Output "Building dev containers with tag $TAG"

        Write-Output "Building container $DOCKER_PKG/${NAME}-arm-dev:$TAG"
        Invoke-Checked "docker build -f dockerfile-arm --platform linux/arm/v7 --build-arg NAME=${NAME}-arm -t $DOCKER_PKG/${NAME}-arm-dev:$TAG dist"
    }
    elseif ($TAG -match $PRD_TAG_REGEX) {
        Write-Output "Building prd containers with tag $TAG"

        Write-Output "Building container $DOCKER_PKG/${NAME}-arm:$TAG"
        Invoke-Checked "docker build -f dockerfile-arm --platform linux/arm/v7 --build-arg NAME=${NAME}-arm -t $DOCKER_PKG/${NAME}-arm:$TAG dist"
    }
    else {
        Write-Output "Docker function quit unexpectedly. Commit Hash: $COMMIT_HASH     Tag: $TAG"
    }
}

function Deploy {
    Write-Output "Deploy      Commit Hash: $COMMIT_HASH     Tag: $TAG"

    Write-Output "Logging into repo"
    Invoke-Checked "docker login $DOCKER_URL -u $Env:DOCKER_USERNAME -p $Env:DOCKER_PASSWORD"

    if ($COMMIT_HASH -eq $TAG) {
        Write-Output "Pushing dev containers with tag $COMMIT_HASH"

        Write-Output "Pushing container $DOCKER_PKG/${NAME}-arm-dev:$COMMIT_HASH"
        Invoke-Checked "docker push $DOCKER_PKG/${NAME}-arm-dev:$COMMIT_HASH"
    }
    elseif ($TAG -match $DEV_TAG_REGEX) {
        Write-Output "Pushing dev containers with tag $TAG"

        Write-Output "Pushing container $DOCKER_PKG/${NAME}-arm-dev:$TAG"
        Invoke-Checked "docker push $DOCKER_PKG/${NAME}-arm-dev:$TAG"
    }
    elseif ($TAG -match $PRD_TAG_REGEX) {
        Write-Output "Pushing prd containers with tag $TAG"

        Write-Output "Pushing container $DOCKER_PKG/${NAME}-arm:$TAG"
        Invoke-Checked "docker push $DOCKER_PKG/${NAME}-arm:$TAG"
    }
    else {
        Write-Output "Deploy function quit unexpectedly. Commit Hash: $COMMIT_HASH     Tag: $TAG"
    }
}

if ($COMMAND -eq "All") {
    Cleanup
    Build
    All
}
elseif ($COMMAND -eq "Test") {
    Deps
    Test
}
elseif ($COMMAND -eq "Test-cov") {
    Deps
    Test-cov
}
elseif ($COMMAND -eq "Lint") {
    Deps
    Lint
}
elseif ($COMMAND -eq "Deps") {
    Deps
}
elseif ($COMMAND -eq "BuildOnly") {
    Build
}
elseif ($COMMAND -eq "Build") {
    Cleanup
    Deps
    Build
}
elseif ($COMMAND -eq "Clean") {
    Cleanup
}
elseif ($COMMAND -eq "DockerOnly") {
    DockerFunc
}
elseif ($COMMAND -eq "Docker" ) {
    Cleanup
    Deps
    Build
    DockerFunc
    Cleanup
}
elseif ($COMMAND -eq "DeployOnly") {
    Deploy
}
elseif ($COMMAND -eq "Deploy" ) {
    Cleanup
    Deps
    Build
    DockerFunc
    Deploy
    Cleanup
}
else {
    Write-Output "Please include a valid command parameter"
}
