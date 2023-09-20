# artifact-deployer

A Go program that listens for HTTP requests and deploys the latest artifact from a GitHub action.

## What does this do?

1. Configure with token, correct paths, etc
2. Run
3. When you `PUT` to `/deploy`, the program will download the latest artifact and unzip it to `~/www/$SITE_DIR`

## Why would you write a program to do this instead of pushing with an action?

- I wanted a project to learn Go
- This way I don't have to give a GitHub action SSH access to anything

## Todo:

- [ ] Check which SHA is deployed to avoid deploying more than once
