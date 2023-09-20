# artifact-deployer

A Go program that listens for HTTP requests and deploys the latest artifact from a GitHub action.

## What does this do?

1. Configure with token, correct paths, etc
2. Run
3. When you `PUT` to `/deploy`, the program will download the latest artifact and unzip it to `~/www/$SITE_DIR`

### PUT parameters

The PUT request requires two parameters:

1. `headsha`: The commit SHA of the artifact you want to deploy.
2. `handshake`: sha1(secret + headsha). This is to make it harder for internet pranksters to try to deploy your artifacts for you.

## Why would you write a program to do this instead of pushing with an action?

- I wanted a project to learn Go
- This way I don't have to give a GitHub action SSH access to anything

## Configuration

This program uses several environment variables for its configuration.

| variable | function | example value |
|----------|----------|---------------|
| WELCOME_MSG | Displayed in console on server start | "Artifact deployer welcome message" |
| SITE_DIR | The directory under `~/www` you want to deploy to | benallen.dev |
| TEMP_FILENAME | The filename used for downloading the artifact. Stored in `/tmp/` and removed after extracting. | archive.zip |
| GITHUB_PAT | A GitHub Personal Access Token used to access the GitHub API | github_pat_your_access_token_here |
| GITHUB_USER | The username for the artifact you want to deploy | benallen-dev |
| GITHUB_REPO | The repository you want to deploy. Must be owned by GITHUB_USER. | benallen-dot-dev |
| DEPLOY_SECRET | The secret used to hash with the commit SHA to discourage shenanigans. Your action will need the same secret to generate correct handshakes. | YourSecretHere |

## Building for use

```
cd artifact-deployer
go build
```

or if you want to install it with your other go binaries

```
cd artifact-deployer
go install
```


## Todo:
- [ ] Check which SHA is deployed to avoid deploying more than once
- [ ] I'm not super happy with how I'm doing code splitting here, having everything in the root as part of the `main` package feels a bit disorganised.
