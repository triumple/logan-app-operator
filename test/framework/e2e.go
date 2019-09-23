package framework

// E2E is the struct for e2e test stage
// 	- Build is for building e2e test context
// 	- Check is for checking result for previous Build
//	- Update is for updating e2e test context
// 	- Recheck is for checking result for previous Update
// 	- BuildAndCheck is composed of Build and Check
//	- UpdateAndCheck is composed of Update and Check
type E2E struct {
	Build          func()
	Check          func()
	Update         func()
	Recheck        func()
	BuildAndCheck  func()
	UpdateAndCheck func()
}

// Run is the entry point for each e2e test
func (e2e *E2E) Run() {
	if e2e.Build != nil {
		e2e.Build()
	}
	if e2e.Check != nil {
		e2e.Check()
	}
	if e2e.Build == nil && e2e.Check == nil {
		if e2e.BuildAndCheck != nil {
			e2e.BuildAndCheck()
		}
	}
	if e2e.Update != nil {
		e2e.Update()
	}
	if e2e.Recheck != nil {
		e2e.Recheck()
	}
	if e2e.Update == nil && e2e.Recheck == nil {
		if e2e.UpdateAndCheck != nil {
			e2e.UpdateAndCheck()
		}
	}
}
