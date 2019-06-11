package v1

import "github.com/logancloud/logan-app-operator/pkg/logan"

// 1. Java
// Boot -> JavaBoot
func (in *Boot) DeepCopyToJava(out *JavaBoot) {
	*out = JavaBoot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// Boot -> JavaBoot
func (in *Boot) DeepCopyJava() *JavaBoot {
	if in == nil {
		return nil
	}
	out := new(JavaBoot)
	in.DeepCopyToJava(out)
	return out
}

// JavaBoot->Boot
func (in *JavaBoot) DeepCopyIntoBoot(out *Boot) {
	*out = Boot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// JavaBoot->Boot
func (in *JavaBoot) DeepCopyBoot() *Boot {
	if in == nil {
		return nil
	}
	out := new(Boot)
	in.DeepCopyIntoBoot(out)

	out.AppKey = logan.JavaAppKey
	out.BootType = logan.BootJava
	return out
}

// 2. PHP
// Boot -> PhpBoot
func (in *Boot) DeepCopyToPhp(out *PhpBoot) {
	*out = PhpBoot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// Boot -> PhpBoot
func (in *Boot) DeepCopyPhp() *PhpBoot {
	if in == nil {
		return nil
	}
	out := new(PhpBoot)
	in.DeepCopyToPhp(out)
	return out
}

// PhpBoot -> Boot
func (in *PhpBoot) DeepCopyIntoBoot(out *Boot) {
	*out = Boot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// PhpBoot -> Boot
func (in *PhpBoot) DeepCopyBoot() *Boot {
	if in == nil {
		return nil
	}
	out := new(Boot)
	in.DeepCopyIntoBoot(out)

	out.AppKey = logan.PhpAppKey
	out.BootType = logan.BootPhp
	return out
}

// 3. Python
// Boot -> PythonBoot
func (in *Boot) DeepCopyToPython(out *PythonBoot) {
	*out = PythonBoot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// Boot -> PythonBoot
func (in *Boot) DeepCopyPython() *PythonBoot {
	if in == nil {
		return nil
	}
	out := new(PythonBoot)
	in.DeepCopyToPython(out)
	return out
}

// PythonBoot -> Boot
func (in *PythonBoot) DeepCopyIntoBoot(out *Boot) {
	*out = Boot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// PythonBoot -> Boot
func (in *PythonBoot) DeepCopyBoot() *Boot {
	if in == nil {
		return nil
	}
	out := new(Boot)
	in.DeepCopyIntoBoot(out)

	out.AppKey = logan.PythonAppKey
	out.BootType = logan.BootPython

	return out
}

// 4. NodeJS
// Boot -> NodeJSBoot
func (in *Boot) DeepCopyToNodeJS(out *NodeJSBoot) {
	*out = NodeJSBoot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// Boot -> NodeJSBoot
func (in *Boot) DeepCopyNodeJS() *NodeJSBoot {
	if in == nil {
		return nil
	}
	out := new(NodeJSBoot)
	in.DeepCopyToNodeJS(out)
	return out
}

// NodeJSBoot -> Boot
func (in *NodeJSBoot) DeepCopyIntoBoot(out *Boot) {
	*out = Boot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// NodeJSBoot -> Boot
func (in *NodeJSBoot) DeepCopyBoot() *Boot {
	if in == nil {
		return nil
	}
	out := new(Boot)
	in.DeepCopyIntoBoot(out)

	out.AppKey = logan.NodeJSAppKey
	out.BootType = logan.BootNodeJS
	return out
}

// 5. Web
// Boot -> WebBoot
func (in *Boot) DeepCopyToWeb(out *WebBoot) {
	*out = WebBoot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// Boot -> WebBoot
func (in *Boot) DeepCopyWeb() *WebBoot {
	if in == nil {
		return nil
	}
	out := new(WebBoot)
	in.DeepCopyToWeb(out)
	return out
}

// WebBoot -> Boot
func (in *WebBoot) DeepCopyIntoBoot(out *Boot) {
	*out = Boot{}
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// WebBoot -> Boot
func (in *WebBoot) DeepCopyBoot() *Boot {
	if in == nil {
		return nil
	}
	out := new(Boot)
	in.DeepCopyIntoBoot(out)

	out.AppKey = logan.WebAppKey
	out.BootType = logan.BootWeb
	return out
}
