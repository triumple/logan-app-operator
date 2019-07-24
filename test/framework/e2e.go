package framework

type E2E struct {
	Build          func()
	Check          func()
	Update         func()
	Recheck        func()
	BuildAndCheck  func()
	UpdateAndCheck func()
}

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
