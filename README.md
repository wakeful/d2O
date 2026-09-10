# d2O

> **d**elete **2** **O**r more!

> [!NOTE]
> `d2O` is a small CLI that audits your AWS IAM **customer-managed policies** and reports two things:
>
> 1. **Unused policies** - policies not attached to any principal and not used as a permissions boundary (attachment count and permissions-boundary usage count are both `0`).
> 2. **Policies with multiple versions** - non-default versions that exist alongside the active one. IAM allows at most **5 versions** per managed policy, stale versions are a security risk as they may hold higher permissions than the current default.

```shell
$ ./d2O -h
Usage of d2O:
  -region string
    	AWS region used for IAM communication (default "eu-west-1")
  -verbose
    	verbose log output
  -version
    	show version
```

### Installation

#### From source

```shell
# via the Go toolchain
go install github.com/wakeful/d2O@latest
```

#### Using a binary release

You can download a pre-built binary from the [release page](https://github.com/wakeful/d2O/releases/latest) and add it
to your user PATH.

### Example scenario

Let's run `d2O` against the current AWS account.

```shell
$ d2O | tee output
```

We should get back a similar response.

```json
{
  "unused": [
    "arn:aws:iam::CurrentAccountID:policy/test-admin",
  ],
  "withVersions": {
    "arn:aws:iam::CurrentAccountID:policy/backup": [
      {
        "createDate": "2026-04-08T18:42:00Z",
        "versionID": "v3"
      },
      {
        "createDate": "2026-04-08T18:24:00Z",
        "versionID": "v1"
      }
    ]
  }
}
```
