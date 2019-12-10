# Gogen

## Publish releases
For a release the tool `goreleaser` is used. You can get it [here](https://github.com/goreleaser/goreleaser/releases). Download your needed binary file and extract to your `$GOPATH/bin`. `$GOPATH/bin` should be on your path environment variables.

Also export your VCS token like: `GITHUB_TOKEN, GITLAB_TOKEN and GITEA_TOKEN`

`goreleaser` builds from the latest tag which you made and it requires a clean working copy. Then `goreleaser` places the build output into a folder ``dist``

