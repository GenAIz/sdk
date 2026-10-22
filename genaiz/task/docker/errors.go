package docker

import (
	"genaiz.com/genaiz/task"
)

var (
	Errors = &dockerErrors{}
)

type dockerErrors struct {
	task.EagerErrors
	buildImageNotFound            task.Error
	buildImageRebuildRequired     task.Error
	containerNameRequired         task.Error
	containerNotFound             task.Error
	listImageNotFound             task.Error
	pushImageNoBuild              task.Error
	pushImageNoProvision          task.Error
	pushImageConflict             task.Error
	pushImageInactive             task.Error
	pushImageNeutral              task.Error
	pushImageUnknownRepoDigest    task.Error
	pushImageUnrecognizedResponse task.Error
	runImageNotFound              task.Error
}

func (e *dockerErrors) BuildImageNotFound() error {
	return e.SetOrReturn(&e.buildImageNotFound,
		task.MakeErrorProvider("could not find a build image"))
}

func (e *dockerErrors) BuildImageRebuildRequired() error {
	return e.SetOrReturn(&e.buildImageRebuildRequired,
		task.MakeErrorProvider("latest needs to be rebuilt"))
}

func (e *dockerErrors) ContainerNameRequired() error {
	return e.SetOrReturn(&e.containerNameRequired,
		task.MakeErrorProvider("a container name could not be elected"))
}

func (e *dockerErrors) ContainerNotFound() error {
	return e.SetOrReturn(&e.containerNotFound,
		task.MakeErrorProvider("container not found"))
}

func (e *dockerErrors) ListImageNotFound(reference string) error {
	var parent = e.SetOrReturn(&e.listImageNotFound,
		task.MakeErrorProvider("could not find image reference"))

	return task.NewFormattedError(parent, "could not find image [%s]", reference)
}

func (e *dockerErrors) PushImageConflict() error {
	return e.SetOrReturn(&e.pushImageConflict,
		task.MakeErrorProvider("function can not be provisioned"))
}

func (e *dockerErrors) PushImageInactive() error {
	return e.SetOrReturn(&e.pushImageInactive,
		task.MakeErrorProvider("can not push function, it is inactive"))
}

func (e *dockerErrors) PushImageNeutral() error {
	return e.SetOrReturn(&e.pushImageNeutral,
		task.MakeErrorProvider("function digest is already known"))
}

func (e *dockerErrors) PushImageNoBuild() error {
	return e.SetOrReturn(&e.pushImageNoBuild,
		task.MakeErrorProvider("function push called without an image built"))
}

func (e *dockerErrors) PushImageNoProvision() error {
	return e.SetOrReturn(&e.pushImageNoProvision,
		task.MakeErrorProvider("function push called without a provision"))
}

func (e *dockerErrors) PushImageUnrecognizedResponse() error {
	return e.SetOrReturn(&e.pushImageUnrecognizedResponse,
		task.MakeErrorProvider("docker push failed without a proper response"))
}

func (e *dockerErrors) PushImageUnknownRepoDigest() error {
	return e.SetOrReturn(&e.pushImageUnknownRepoDigest,
		task.MakeErrorProvider("function repository digest is unknown"))
}

func (e *dockerErrors) RunImageNotFound() error {
	return e.SetOrReturn(&e.runImageNotFound,
		task.MakeErrorProvider("could not find an image to run"))
}
