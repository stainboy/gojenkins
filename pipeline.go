package gojenkins

import (
	"errors"
	"fmt"
)

type Pipeline struct {
	jenkins *Jenkins
	job     *Job
	build   *Build
	raw     *PipelineBody
}

type PipelineBody struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	StartTimeMillis     int64  `json:"startTimeMillis"`
	EndTimeMillis       int64  `json:"endTimeMillis"`
	DurationMillis      int64  `json:"durationMillis"`
	QueueDurationMillis int64  `json:"queueDurationMillis"`
	PauseDurationMillis int64  `json:"pauseDurationMillis"`
	Stages              []struct {
		ID                  int64  `json:"id"`
		Name                string `json:"name"`
		ExecNode            string `json:"execNode"`
		Status              string `json:"status"`
		StartTimeMillis     int64  `json:"startTimeMillis"`
		DurationMillis      int64  `json:"durationMillis"`
		PauseDurationMillis int64  `json:"pauseDurationMillis"`
	} `json:"stages"`
}

func (p *Pipeline) Poll() (int, error) {

	if p.raw == nil {
		p.raw = new(PipelineBody)
	}

	url := fmt.Sprintf("/job/%s/%d/wfapi/describe", p.job.GetName(), p.build.GetBuildNumber())
	response, err := p.jenkins.Requester.GetJSON(url, p.raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}

func (p *Pipeline) GetRaw() (*PipelineBody, error) {
	if p.raw == nil {
		return nil, errors.New("Pipeline body is nil")
	}
	return p.raw, nil
}
