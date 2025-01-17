// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/tanzu-cli/cmd/plugin/builder/imgpkg"
)

type ImgpkgWrapper struct {
	CopyArchiveToRepoStub        func(string, string) error
	copyArchiveToRepoMutex       sync.RWMutex
	copyArchiveToRepoArgsForCall []struct {
		arg1 string
		arg2 string
	}
	copyArchiveToRepoReturns struct {
		result1 error
	}
	copyArchiveToRepoReturnsOnCall map[int]struct {
		result1 error
	}
	CopyImageToArchiveStub        func(string, string) error
	copyImageToArchiveMutex       sync.RWMutex
	copyImageToArchiveArgsForCall []struct {
		arg1 string
		arg2 string
	}
	copyImageToArchiveReturns struct {
		result1 error
	}
	copyImageToArchiveReturnsOnCall map[int]struct {
		result1 error
	}
	GetFileDigestFromImageStub        func(string, string) (string, error)
	getFileDigestFromImageMutex       sync.RWMutex
	getFileDigestFromImageArgsForCall []struct {
		arg1 string
		arg2 string
	}
	getFileDigestFromImageReturns struct {
		result1 string
		result2 error
	}
	getFileDigestFromImageReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	PullImageStub        func(string, string) error
	pullImageMutex       sync.RWMutex
	pullImageArgsForCall []struct {
		arg1 string
		arg2 string
	}
	pullImageReturns struct {
		result1 error
	}
	pullImageReturnsOnCall map[int]struct {
		result1 error
	}
	PushImageStub        func(string, string) error
	pushImageMutex       sync.RWMutex
	pushImageArgsForCall []struct {
		arg1 string
		arg2 string
	}
	pushImageReturns struct {
		result1 error
	}
	pushImageReturnsOnCall map[int]struct {
		result1 error
	}
	ResolveImageStub        func(string) error
	resolveImageMutex       sync.RWMutex
	resolveImageArgsForCall []struct {
		arg1 string
	}
	resolveImageReturns struct {
		result1 error
	}
	resolveImageReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ImgpkgWrapper) CopyArchiveToRepo(arg1 string, arg2 string) error {
	fake.copyArchiveToRepoMutex.Lock()
	ret, specificReturn := fake.copyArchiveToRepoReturnsOnCall[len(fake.copyArchiveToRepoArgsForCall)]
	fake.copyArchiveToRepoArgsForCall = append(fake.copyArchiveToRepoArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.CopyArchiveToRepoStub
	fakeReturns := fake.copyArchiveToRepoReturns
	fake.recordInvocation("CopyArchiveToRepo", []interface{}{arg1, arg2})
	fake.copyArchiveToRepoMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ImgpkgWrapper) CopyArchiveToRepoCallCount() int {
	fake.copyArchiveToRepoMutex.RLock()
	defer fake.copyArchiveToRepoMutex.RUnlock()
	return len(fake.copyArchiveToRepoArgsForCall)
}

func (fake *ImgpkgWrapper) CopyArchiveToRepoCalls(stub func(string, string) error) {
	fake.copyArchiveToRepoMutex.Lock()
	defer fake.copyArchiveToRepoMutex.Unlock()
	fake.CopyArchiveToRepoStub = stub
}

func (fake *ImgpkgWrapper) CopyArchiveToRepoArgsForCall(i int) (string, string) {
	fake.copyArchiveToRepoMutex.RLock()
	defer fake.copyArchiveToRepoMutex.RUnlock()
	argsForCall := fake.copyArchiveToRepoArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ImgpkgWrapper) CopyArchiveToRepoReturns(result1 error) {
	fake.copyArchiveToRepoMutex.Lock()
	defer fake.copyArchiveToRepoMutex.Unlock()
	fake.CopyArchiveToRepoStub = nil
	fake.copyArchiveToRepoReturns = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) CopyArchiveToRepoReturnsOnCall(i int, result1 error) {
	fake.copyArchiveToRepoMutex.Lock()
	defer fake.copyArchiveToRepoMutex.Unlock()
	fake.CopyArchiveToRepoStub = nil
	if fake.copyArchiveToRepoReturnsOnCall == nil {
		fake.copyArchiveToRepoReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.copyArchiveToRepoReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) CopyImageToArchive(arg1 string, arg2 string) error {
	fake.copyImageToArchiveMutex.Lock()
	ret, specificReturn := fake.copyImageToArchiveReturnsOnCall[len(fake.copyImageToArchiveArgsForCall)]
	fake.copyImageToArchiveArgsForCall = append(fake.copyImageToArchiveArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.CopyImageToArchiveStub
	fakeReturns := fake.copyImageToArchiveReturns
	fake.recordInvocation("CopyImageToArchive", []interface{}{arg1, arg2})
	fake.copyImageToArchiveMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ImgpkgWrapper) CopyImageToArchiveCallCount() int {
	fake.copyImageToArchiveMutex.RLock()
	defer fake.copyImageToArchiveMutex.RUnlock()
	return len(fake.copyImageToArchiveArgsForCall)
}

func (fake *ImgpkgWrapper) CopyImageToArchiveCalls(stub func(string, string) error) {
	fake.copyImageToArchiveMutex.Lock()
	defer fake.copyImageToArchiveMutex.Unlock()
	fake.CopyImageToArchiveStub = stub
}

func (fake *ImgpkgWrapper) CopyImageToArchiveArgsForCall(i int) (string, string) {
	fake.copyImageToArchiveMutex.RLock()
	defer fake.copyImageToArchiveMutex.RUnlock()
	argsForCall := fake.copyImageToArchiveArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ImgpkgWrapper) CopyImageToArchiveReturns(result1 error) {
	fake.copyImageToArchiveMutex.Lock()
	defer fake.copyImageToArchiveMutex.Unlock()
	fake.CopyImageToArchiveStub = nil
	fake.copyImageToArchiveReturns = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) CopyImageToArchiveReturnsOnCall(i int, result1 error) {
	fake.copyImageToArchiveMutex.Lock()
	defer fake.copyImageToArchiveMutex.Unlock()
	fake.CopyImageToArchiveStub = nil
	if fake.copyImageToArchiveReturnsOnCall == nil {
		fake.copyImageToArchiveReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.copyImageToArchiveReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) GetFileDigestFromImage(arg1 string, arg2 string) (string, error) {
	fake.getFileDigestFromImageMutex.Lock()
	ret, specificReturn := fake.getFileDigestFromImageReturnsOnCall[len(fake.getFileDigestFromImageArgsForCall)]
	fake.getFileDigestFromImageArgsForCall = append(fake.getFileDigestFromImageArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.GetFileDigestFromImageStub
	fakeReturns := fake.getFileDigestFromImageReturns
	fake.recordInvocation("GetFileDigestFromImage", []interface{}{arg1, arg2})
	fake.getFileDigestFromImageMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ImgpkgWrapper) GetFileDigestFromImageCallCount() int {
	fake.getFileDigestFromImageMutex.RLock()
	defer fake.getFileDigestFromImageMutex.RUnlock()
	return len(fake.getFileDigestFromImageArgsForCall)
}

func (fake *ImgpkgWrapper) GetFileDigestFromImageCalls(stub func(string, string) (string, error)) {
	fake.getFileDigestFromImageMutex.Lock()
	defer fake.getFileDigestFromImageMutex.Unlock()
	fake.GetFileDigestFromImageStub = stub
}

func (fake *ImgpkgWrapper) GetFileDigestFromImageArgsForCall(i int) (string, string) {
	fake.getFileDigestFromImageMutex.RLock()
	defer fake.getFileDigestFromImageMutex.RUnlock()
	argsForCall := fake.getFileDigestFromImageArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ImgpkgWrapper) GetFileDigestFromImageReturns(result1 string, result2 error) {
	fake.getFileDigestFromImageMutex.Lock()
	defer fake.getFileDigestFromImageMutex.Unlock()
	fake.GetFileDigestFromImageStub = nil
	fake.getFileDigestFromImageReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *ImgpkgWrapper) GetFileDigestFromImageReturnsOnCall(i int, result1 string, result2 error) {
	fake.getFileDigestFromImageMutex.Lock()
	defer fake.getFileDigestFromImageMutex.Unlock()
	fake.GetFileDigestFromImageStub = nil
	if fake.getFileDigestFromImageReturnsOnCall == nil {
		fake.getFileDigestFromImageReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getFileDigestFromImageReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *ImgpkgWrapper) PullImage(arg1 string, arg2 string) error {
	fake.pullImageMutex.Lock()
	ret, specificReturn := fake.pullImageReturnsOnCall[len(fake.pullImageArgsForCall)]
	fake.pullImageArgsForCall = append(fake.pullImageArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.PullImageStub
	fakeReturns := fake.pullImageReturns
	fake.recordInvocation("PullImage", []interface{}{arg1, arg2})
	fake.pullImageMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ImgpkgWrapper) PullImageCallCount() int {
	fake.pullImageMutex.RLock()
	defer fake.pullImageMutex.RUnlock()
	return len(fake.pullImageArgsForCall)
}

func (fake *ImgpkgWrapper) PullImageCalls(stub func(string, string) error) {
	fake.pullImageMutex.Lock()
	defer fake.pullImageMutex.Unlock()
	fake.PullImageStub = stub
}

func (fake *ImgpkgWrapper) PullImageArgsForCall(i int) (string, string) {
	fake.pullImageMutex.RLock()
	defer fake.pullImageMutex.RUnlock()
	argsForCall := fake.pullImageArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ImgpkgWrapper) PullImageReturns(result1 error) {
	fake.pullImageMutex.Lock()
	defer fake.pullImageMutex.Unlock()
	fake.PullImageStub = nil
	fake.pullImageReturns = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) PullImageReturnsOnCall(i int, result1 error) {
	fake.pullImageMutex.Lock()
	defer fake.pullImageMutex.Unlock()
	fake.PullImageStub = nil
	if fake.pullImageReturnsOnCall == nil {
		fake.pullImageReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.pullImageReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) PushImage(arg1 string, arg2 string) error {
	fake.pushImageMutex.Lock()
	ret, specificReturn := fake.pushImageReturnsOnCall[len(fake.pushImageArgsForCall)]
	fake.pushImageArgsForCall = append(fake.pushImageArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.PushImageStub
	fakeReturns := fake.pushImageReturns
	fake.recordInvocation("PushImage", []interface{}{arg1, arg2})
	fake.pushImageMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ImgpkgWrapper) PushImageCallCount() int {
	fake.pushImageMutex.RLock()
	defer fake.pushImageMutex.RUnlock()
	return len(fake.pushImageArgsForCall)
}

func (fake *ImgpkgWrapper) PushImageCalls(stub func(string, string) error) {
	fake.pushImageMutex.Lock()
	defer fake.pushImageMutex.Unlock()
	fake.PushImageStub = stub
}

func (fake *ImgpkgWrapper) PushImageArgsForCall(i int) (string, string) {
	fake.pushImageMutex.RLock()
	defer fake.pushImageMutex.RUnlock()
	argsForCall := fake.pushImageArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ImgpkgWrapper) PushImageReturns(result1 error) {
	fake.pushImageMutex.Lock()
	defer fake.pushImageMutex.Unlock()
	fake.PushImageStub = nil
	fake.pushImageReturns = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) PushImageReturnsOnCall(i int, result1 error) {
	fake.pushImageMutex.Lock()
	defer fake.pushImageMutex.Unlock()
	fake.PushImageStub = nil
	if fake.pushImageReturnsOnCall == nil {
		fake.pushImageReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.pushImageReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) ResolveImage(arg1 string) error {
	fake.resolveImageMutex.Lock()
	ret, specificReturn := fake.resolveImageReturnsOnCall[len(fake.resolveImageArgsForCall)]
	fake.resolveImageArgsForCall = append(fake.resolveImageArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ResolveImageStub
	fakeReturns := fake.resolveImageReturns
	fake.recordInvocation("ResolveImage", []interface{}{arg1})
	fake.resolveImageMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ImgpkgWrapper) ResolveImageCallCount() int {
	fake.resolveImageMutex.RLock()
	defer fake.resolveImageMutex.RUnlock()
	return len(fake.resolveImageArgsForCall)
}

func (fake *ImgpkgWrapper) ResolveImageCalls(stub func(string) error) {
	fake.resolveImageMutex.Lock()
	defer fake.resolveImageMutex.Unlock()
	fake.ResolveImageStub = stub
}

func (fake *ImgpkgWrapper) ResolveImageArgsForCall(i int) string {
	fake.resolveImageMutex.RLock()
	defer fake.resolveImageMutex.RUnlock()
	argsForCall := fake.resolveImageArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ImgpkgWrapper) ResolveImageReturns(result1 error) {
	fake.resolveImageMutex.Lock()
	defer fake.resolveImageMutex.Unlock()
	fake.ResolveImageStub = nil
	fake.resolveImageReturns = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) ResolveImageReturnsOnCall(i int, result1 error) {
	fake.resolveImageMutex.Lock()
	defer fake.resolveImageMutex.Unlock()
	fake.ResolveImageStub = nil
	if fake.resolveImageReturnsOnCall == nil {
		fake.resolveImageReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.resolveImageReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ImgpkgWrapper) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.copyArchiveToRepoMutex.RLock()
	defer fake.copyArchiveToRepoMutex.RUnlock()
	fake.copyImageToArchiveMutex.RLock()
	defer fake.copyImageToArchiveMutex.RUnlock()
	fake.getFileDigestFromImageMutex.RLock()
	defer fake.getFileDigestFromImageMutex.RUnlock()
	fake.pullImageMutex.RLock()
	defer fake.pullImageMutex.RUnlock()
	fake.pushImageMutex.RLock()
	defer fake.pushImageMutex.RUnlock()
	fake.resolveImageMutex.RLock()
	defer fake.resolveImageMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ImgpkgWrapper) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ imgpkg.ImgpkgWrapper = new(ImgpkgWrapper)
