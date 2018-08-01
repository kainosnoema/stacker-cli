package backend

import (
	"fmt"

	"github.com/eyeamera/stacker-cli/client"
)

type stack struct {
	name          string
	region        string
	capabilities  string
	templateBody  string
	rawParameters RawParams
	resolver      ParamsResolver
}

func (s *stack) Name() string         { return s.name }
func (s *stack) TemplateBody() string { return s.templateBody }
func (s *stack) Capabilities() []string {
	if s.capabilities != "" {
		return []string{s.capabilities}
	}
	return nil
}
func (s *stack) Region() string { return s.region }
func (s *stack) Params() (client.StackParams, error) {
	return s.resolver.Resolve(s.rawParameters, s)
}

type fetcher struct {
	cs ConfigStore
	ts TemplateStore
	r  ParamsResolver
}

func newFetcher(cs ConfigStore, ts TemplateStore, r ParamsResolver) *fetcher {
	return &fetcher{cs, ts, r}
}

func (f *fetcher) Fetch(name string) ([]client.Stack, error) {
	stacks := make([]client.Stack, 0)
	stackConfigs, err := f.cs.Fetch(name)
	if err != nil {
		return stacks, fmt.Errorf("unable to fetch stack %s: %s", name, err)
	}

	// Fetch the templates for each stack to get a final list of params,
	// and the template body
	for _, stackConfig := range stackConfigs {
		t, err := f.ts.Fetch(stackConfig.TemplateName)
		if err != nil {
			return stacks, fmt.Errorf("unable to fetch template %s: %s", stackConfig.TemplateName, err)
		}

		rp := make(RawParams)
		for _, k := range t.Parameters() {
			if v, ok := stackConfig.Parameters[k]; ok {
				rp[k] = v
			}
		}

		s := &stack{
			name:          stackConfig.Name,
			region:        stackConfig.Region,
			capabilities:  stackConfig.Capabilities,
			templateBody:  t.Body(),
			rawParameters: rp,
			resolver:      f.r,
		}

		stacks = append(stacks, s)
	}

	return stacks, nil
}