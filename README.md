# Jenkins API Client for Go

Forked from https://github.com/bndr/gojenkins

[![GoDoc](https://godoc.org/github.com/bndr/gojenkins?status.svg)](https://godoc.org/github.com/bndr/gojenkins)

## Installation

    go get github.com/stainboy/gojenkins

## Usage

```go

import "github.com/stainboy/gojenkins"

jenkins, err := gojenkins.CreateJenkins("http://localhost:8080/", "admin", "admin").Init()

if err != nil {
  panic("Something Went Wrong")
}

job, err := jenkins.GetJob("sip")
if err != nil {
  panic("Job Does Not Exist")
}

lastSuccessBuild := job.GetLastSuccessfulBuild()
if err != nil {
  panic("Last SuccessBuild does not exist")
}

duration := lastSuccessBuild.GetDuration()


```

API Reference: https://godoc.org/github.com/bndr/gojenkins

## Examples

For all of the examples below first create a jenkins object
```go
import "github.com/bndr/gojenkins"

jenkins, _ := gojenkins.CreateJenkins("http://localhost:8080/", "admin", "admin").Init()
```

or if you don't need authentication:

```go
jenkins, _ := gojenkins.CreateJenkins("http://localhost:8080/").Init()
```

### Check Status of all nodes

```go
nodes := jenkins.GetAllNodes()

for _, node := range nodes {

  // Fetch Node Data
  node.Poll()
	if node.IsOnline() {
		fmt.Println("Node is Online")
	}
}

```

### Get all Builds for specific Job, and check their status

```go
jobName := "someJob"
builds, err := jenkins.GetAllBuildIds(jobName)

if err != nil {
  panic(err)
}

for _, build := range builds {
  buildId := build.Number
  data, err := jenkins.GetBuild(jobName, buildId)

  if err != nil {
    panic(err)
  }

	if "SUCCESS" == data.GetResult() {
		fmt.Println("This build succeeded")
	}
}

// Get Last Successful/Failed/Stable Build for a Job
job, err := jenkins.GetJob("someJob")

if err != nil {
  panic(err)
}

job.GetLastSuccessfulBuild()
job.GetLastStableBuild()

```

### Get Current Tasks in Queue, and the reason why they're in the queue

```go

tasks := jenkins.GetQueue()

for _, task := range tasks {
	fmt.Println(task.GetWhy())
}

```

### Create View and add Jobs to it

```go

view, err := jenkins.CreateView("test_view", gojenkins.LIST_VIEW)

if err != nil {
  panic(err)
}

status, err := view.AddJob("jobName")

if status {
  fmt.Println("Job has been added to view")
}

```

### Get All Artifacts for a Build and Save them to a folder

```go

job, _ := jenkins.GetJob("job")
build, _ := job.GetBuild(1)
artifacts := build.GetArtifacts()

for _, a := range artifacts {
	a.SaveToDir("/tmp")
}

```

### To always get fresh data use the .Poll() method

```go

job, _ := jenkins.GetJob("job")
job.Poll()

build, _ := job.getBuild(1)
build.Poll()

```

## Testing

    go test

## Contribute

All Contributions are welcome. The todo list is on the bottom of this README. Feel free to send a pull request.

## TODO

Although the basic features are implemented there are many optional features that are on the todo list.

* Kerberos Authentication
* CLI Tool
* Rewrite some (all?) iterators with channels

## LICENSE

Apache License 2.0
