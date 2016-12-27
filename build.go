// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"bytes"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type Build struct {
	Raw     *buildResponse
	Job     *Job
	Jenkins *Jenkins
	Base    string
	Depth   int
}

type Parameter struct {
	Name  string
	Value string
}

type branch struct {
	SHA1 string
	Name string
}

type buildRevision struct {
	SHA1   string   `json:"SHA1"`
	Branch []branch `json:"branch"`
}

type builds struct {
	BuildNumber int64         `json:"buildNumber"`
	BuildResult interface{}   `json:"buildResult"`
	Marked      buildRevision `json:"marked"`
	Revision    buildRevision `json:"revision"`
}

type culprit struct {
	AbsoluteUrl string
	FullName    string
}

type generalObj struct {
	Parameters              []Parameter              `json:"parameters"`
	Causes                  []map[string]interface{} `json:"causes"`
	BuildsByBranchName      map[string]builds        `json:"buildsByBranchName"`
	LastBuiltRevision       buildRevision            `json:"lastBuiltRevision"`
	RemoteUrls              []string                 `json:"remoteUrls"`
	ScmName                 string                   `json:"scmName"`
	MercurialNodeName       string                   `json:"mercurialNodeName"`
	MercurialRevisionNumber string                   `json:"mercurialRevisionNumber"`
	Subdir                  interface{}              `json:"subdir"`
	TotalCount              int64
	UrlName                 string
}

type testResult struct {
	Duration  int64 `json:"duration"`
	Empty     bool  `json:"empty"`
	FailCount int64 `json:"failCount"`
	PassCount int64 `json:"passCount"`
	SkipCount int64 `json:"skipCount"`
	Suites    []struct {
		Cases []struct {
			Age             int64       `json:"age"`
			ClassName       string      `json:"className"`
			Duration        int64       `json:"duration"`
			ErrorDetails    interface{} `json:"errorDetails"`
			ErrorStackTrace interface{} `json:"errorStackTrace"`
			FailedSince     int64       `json:"failedSince"`
			Name            string      `json:"name"`
			Skipped         bool        `json:"skipped"`
			SkippedMessage  interface{} `json:"skippedMessage"`
			Status          string      `json:"status"`
			Stderr          interface{} `json:"stderr"`
			Stdout          interface{} `json:"stdout"`
		} `json:"cases"`
		Duration  int64       `json:"duration"`
		ID        interface{} `json:"id"`
		Name      string      `json:"name"`
		Stderr    interface{} `json:"stderr"`
		Stdout    interface{} `json:"stdout"`
		Timestamp interface{} `json:"timestamp"`
	} `json:"suites"`
}

type buildResponse struct {
	Type      string `json:"_class"`
	Actions   []generalObj
	Artifacts []struct {
		DisplayPath  string `json:"displayPath"`
		FileName     string `json:"fileName"`
		RelativePath string `json:"relativePath"`
	} `json:"artifacts"`
	Building  bool   `json:"building"`
	BuiltOn   string `json:"builtOn"`
	ChangeSet struct {
		Items []struct {
			AffectedPaths []string `json:"affectedPaths"`
			Author        struct {
				AbsoluteUrl string `json:"absoluteUrl"`
				FullName    string `json:"fullName"`
			} `json:"author"`
			Comment  string `json:"comment"`
			CommitId string `json:"commitId"`
			Date     string `json:"date"`
			ID       string `json:"id"`
			Msg      string `json:"msg"`
			Paths    []struct {
				EditType string `json:"editType"`
				File     string `json:"file"`
			} `json:"paths"`
			Timestamp int64 `json:"timestamp"`
		} `json:"items"`
		Kind      string `json:"kind"`
		Revisions []struct {
			Module   string
			Revision int
		} `json:"revision"`
	} `json:"changeSet"`
	Culprits          []culprit   `json:"culprits"`
	Description       interface{} `json:"description"`
	Duration          int64       `json:"duration"`
	EstimatedDuration int64       `json:"estimatedDuration"`
	Executor          interface{} `json:"executor"`
	FullDisplayName   string      `json:"fullDisplayName"`
	ID                string      `json:"id"`
	KeepLog           bool        `json:"keepLog"`
	Number            int64       `json:"number"`
	Result            string      `json:"result"`
	Timestamp         int64       `json:"timestamp"`
	URL               string      `json:"url"`
	MavenArtifacts    interface{} `json:"mavenArtifacts"`
	MavenVersionUsed  string      `json:"mavenVersionUsed"`
	Fingerprint       []fingerPrintResponse
	Runs              []struct {
		Number int64
		Url    string
	} `json:"runs"`
}

// Builds
func (b *Build) Info() *buildResponse {
	return b.Raw
}

func (b *Build) GetActions() []generalObj {
	return b.Raw.Actions
}

func (b *Build) GetUrl() string {
	return b.Raw.URL
}

func (b *Build) GetBuildNumber() int64 {
	return b.Raw.Number
}
func (b *Build) GetResult() string {
	return b.Raw.Result
}

func (b *Build) GetArtifacts() []Artifact {
	artifacts := make([]Artifact, len(b.Raw.Artifacts))
	for i, artifact := range b.Raw.Artifacts {
		artifacts[i] = Artifact{
			Jenkins:  b.Jenkins,
			Build:    b,
			FileName: artifact.FileName,
			Path:     b.Base + "/artifact/" + artifact.RelativePath,
		}
	}
	return artifacts
}

func (b *Build) GetCulprits() []culprit {
	return b.Raw.Culprits
}

func (b *Build) Stop() (bool, error) {
	if b.IsRunning() {
		response, err := b.Jenkins.Requester.Post(b.Base+"/stop", nil, nil, nil)
		if err != nil {
			return false, err
		}
		return response.StatusCode == 200, nil
	}
	return true, nil
}

func (b *Build) GetConsoleOutput() string {
	url := b.Base + "/consoleText"
	var content string
	b.Jenkins.Requester.GetXML(url, &content, nil)
	return content
}

func (b *Build) GetCauses() ([]map[string]interface{}, error) {
	for _, a := range b.Raw.Actions {
		if a.Causes != nil {
			return a.Causes, nil
		}
	}
	return nil, errors.New("No Causes")
}

func (b *Build) GetParameters() []Parameter {
	for _, a := range b.Raw.Actions {
		if a.Parameters != nil {
			return a.Parameters
		}
	}
	return nil
}

func (b *Build) GetInjectedEnvVars() (map[string]string, error) {
	var envVars struct {
		EnvMap map[string]string `json:"envMap"`
	}
	endpoint := b.Base + "/injectedEnvVars"
	_, err := b.Jenkins.Requester.GetJSON(endpoint, &envVars, nil)
	if err != nil {
		return envVars.EnvMap, err
	}
	return envVars.EnvMap, nil
}

func (b *Build) GetDownstreamBuilds() ([]*Build, error) {
	downstreamJobs := b.GetDownstreamJobNames()
	fingerprints := b.GetAllFingerprints()
	result := make([]*Build, 0)
	for _, fingerprint := range fingerprints {
		for _, usage := range fingerprint.Raw.Usage {
			if inSlice(usage.Name, downstreamJobs) {
				job, err := b.Jenkins.GetJob(usage.Name)
				if err != nil {
					return nil, err
				}
				for _, ranges := range usage.Ranges.Ranges {
					for i := ranges.Start; i <= ranges.End; i++ {
						build, err := job.GetBuild(i)
						if err != nil {
							return nil, err
						}
						result = append(result, build)
					}
				}
			}
		}
	}
	return result, nil
}

func (b *Build) GetDownstreamJobNames() []string {
	result := make([]string, 0)
	downstreamJobs := b.Job.GetDownstreamJobsMetadata()
	fingerprints := b.GetAllFingerprints()
	for _, fingerprint := range fingerprints {
		for _, usage := range fingerprint.Raw.Usage {
			for _, job := range downstreamJobs {
				if job.Name == usage.Name {
					result = append(result, job.Name)
				}
			}
		}
	}
	return result
}

func (b *Build) GetAllFingerprints() []*Fingerprint {
	b.Poll(3)
	result := make([]*Fingerprint, len(b.Raw.Fingerprint))
	for i, f := range b.Raw.Fingerprint {
		result[i] = &Fingerprint{Jenkins: b.Jenkins, Base: "/fingerprint/", Id: f.Hash, Raw: &f}
	}
	return result
}

func (b *Build) GetUpstreamJob() (*Job, error) {
	causes, err := b.GetCauses()
	if err != nil {
		return nil, err
	}
	if len(causes) > 0 {
		if job, ok := causes[0]["upstreamProject"]; ok {
			return b.Jenkins.GetJob(job.(string))
		}
	}
	return nil, errors.New("Unable to get Upstream Job")
}

func (b *Build) GetUpstreamBuildNumber() (int64, error) {
	causes, err := b.GetCauses()
	if err != nil {
		return 0, err
	}
	if len(causes) > 0 {
		if build, ok := causes[0]["upstreamBuild"]; ok {
			switch t := build.(type) {
			default:
				return t.(int64), nil
			case float64:
				return int64(t), nil
			}
		}
	}
	return 0, nil
}

func (b *Build) GetUpstreamBuild() (*Build, error) {
	job, err := b.GetUpstreamJob()
	if err != nil {
		return nil, err
	}
	if job != nil {
		buildNumber, err := b.GetUpstreamBuildNumber()
		if err == nil {
			return job.GetBuild(buildNumber)
		}
	}
	return nil, errors.New("Build not found")
}

func (b *Build) GetMatrixRuns() ([]*Build, error) {
	_, err := b.Poll(0)
	if err != nil {
		return nil, err
	}
	runs := b.Raw.Runs
	result := make([]*Build, len(b.Raw.Runs))
	r, _ := regexp.Compile("job/(.*?)/(.*?)/(\\d+)/")

	for i, run := range runs {
		result[i] = &Build{Jenkins: b.Jenkins, Job: b.Job, Raw: new(buildResponse), Depth: 1, Base: "/" + r.FindString(run.Url)}
		result[i].Poll()
	}
	return result, nil
}

func (b *Build) GetResultSet() (*testResult, error) {

	url := b.Base + "/testReport"
	var report testResult

	_, err := b.Jenkins.Requester.GetJSON(url, &report, nil)
	if err != nil {
		return nil, err
	}

	return &report, nil

}

func (b *Build) GetTimestamp() time.Time {
	msInt := int64(b.Raw.Timestamp)
	return time.Unix(0, msInt*int64(time.Millisecond))
}

func (b *Build) GetDuration() int64 {
	return b.Raw.Duration
}

func (b *Build) GetRevision() string {
	vcs := b.Raw.ChangeSet.Kind

	if vcs == "git" || vcs == "hg" {
		for _, a := range b.Raw.Actions {
			if a.LastBuiltRevision.SHA1 != "" {
				return a.LastBuiltRevision.SHA1
			}
			if a.MercurialRevisionNumber != "" {
				return a.MercurialRevisionNumber
			}
		}
	} else if vcs == "svn" {
		return strconv.Itoa(b.Raw.ChangeSet.Revisions[0].Revision)
	}
	return ""
}

func (b *Build) GetRevisionBranch() string {
	vcs := b.Raw.ChangeSet.Kind
	if vcs == "git" {
		for _, a := range b.Raw.Actions {
			if len(a.LastBuiltRevision.Branch) > 0 && a.LastBuiltRevision.Branch[0].SHA1 != "" {
				return a.LastBuiltRevision.Branch[0].SHA1
			}
		}
	} else {
		panic("Not implemented")
	}
	return ""
}

func (b *Build) IsGood() bool {
	return (!b.IsRunning() && b.Raw.Result == STATUS_SUCCESS)
}

func (b *Build) IsRunning() bool {
	_, err := b.Poll()
	if err != nil {
		return false
	}
	return b.Raw.Building
}

func (b *Build) GetType() string {
	return b.Raw.Type
}

func (b *Build) GetPipeline() (*Pipeline, error) {
	pipeline := Pipeline{jenkins: b.Jenkins, job: b.Job, build: b}
	status, err := pipeline.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &pipeline, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (b *Build) IsPipeline() bool {
	return b.Raw.Type == "org.jenkinsci.plugins.workflow.job.WorkflowRun"
}

func (b *Build) SetDescription(description string) error {
	data := url.Values{}
	data.Set("description", description)
	if _, err := b.Jenkins.Requester.Post(b.Base+"/submitDescription", bytes.NewBufferString(data.Encode()), nil, nil); err != nil {
		return err
	}

	return nil
}

// Poll for current data. Optional parameter - depth.
// More about depth here: https://wiki.jenkins-ci.org/display/JENKINS/Remote+access+API
func (b *Build) Poll(options ...interface{}) (int, error) {
	depth := "-1"

	for _, o := range options {
		switch v := o.(type) {
		case string:
			depth = v
		case int:
			depth = strconv.Itoa(v)
		case int64:
			depth = strconv.FormatInt(v, 10)
		}
	}
	if depth == "-1" {
		depth = strconv.Itoa(b.Depth)
	}

	qr := map[string]string{
		"depth": depth,
	}
	response, err := b.Jenkins.Requester.GetJSON(b.Base, b.Raw, qr)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
