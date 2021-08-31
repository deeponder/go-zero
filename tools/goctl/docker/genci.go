package docker

import (
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util"
)

const (
	ciTemplate = `services:
  - docker:dind

stages:
  - build

before_script:
  - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY

docker-build:
  stage: build
  script:
    - docker pull $CI_REGISTRY_IMAGE:latest || true

    - docker build --cache-from $CI_REGISTRY_IMAGE:latest --tag $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_SLUG$CI_COMMIT_SHORT_SHA --tag $CI_REGISTRY_IMAGE:latest .

    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_SLUG$CI_COMMIT_SHORT_SHA
    - docker push $CI_REGISTRY_IMAGE:latest
  only:
    - master
  tags:
    - global-runner
`
)

func generateGitlabCifile() error {

	text, err := util.LoadTemplate(category, ciTemplateFile, ciTemplate)
	if err != nil {
		return err
	}

	return util.With("ci").Parse(text).SaveTo(nil, ".gitlab-ci.yml", false)
}
