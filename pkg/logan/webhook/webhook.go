package webhook

import (
	v1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
)

const (
	ApiTypeJava   = "JavaBoot"
	ApiTypePhp    = "PhpBoot"
	ApiTypePython = "PythonBoot"
	ApiTypeNodeJS = "NodeJSBoot"
	ApiTypeWeb    = "WebBoot"
)

// DecodeBoot decode the Boot object from request.
func DecodeBoot(req types.Request, decoder types.Decoder) (*appv1.Boot, error) {
	bootType := req.AdmissionRequest.Kind.Kind

	var boot *appv1.Boot
	if bootType == ApiTypeJava {
		apiBoot := &v1.JavaBoot{}
		err := decoder.Decode(req, apiBoot)
		if err != nil {
			return nil, err
		}
		boot = apiBoot.DeepCopyBoot()
	} else if bootType == ApiTypePhp {
		apiBoot := &v1.PhpBoot{}
		err := decoder.Decode(req, apiBoot)
		if err != nil {
			return nil, err
		}
		boot = apiBoot.DeepCopyBoot()
	} else if bootType == ApiTypePython {
		apiBoot := &v1.PythonBoot{}
		err := decoder.Decode(req, apiBoot)
		if err != nil {
			return nil, err
		}
		boot = apiBoot.DeepCopyBoot()
	} else if bootType == ApiTypeNodeJS {
		apiBoot := &v1.NodeJSBoot{}
		err := decoder.Decode(req, apiBoot)
		if err != nil {
			return nil, err
		}
		boot = apiBoot.DeepCopyBoot()
	} else if bootType == ApiTypeWeb {
		apiBoot := &v1.WebBoot{}
		err := decoder.Decode(req, apiBoot)
		if err != nil {
			return nil, err
		}
		boot = apiBoot.DeepCopyBoot()
	}

	return boot, nil
}

// DecodeJavaBoot decode the JavaBoot object from request.
func DecodeJavaBoot(req types.Request, decoder types.Decoder) (*appv1.JavaBoot, error) {
	bootType := req.AdmissionRequest.Kind.Kind

	var boot *appv1.JavaBoot
	if bootType == ApiTypeJava {
		boot = &v1.JavaBoot{}
		err := decoder.Decode(req, boot)
		if err != nil {
			return nil, err
		}
		return boot, nil
	}

	return boot, nil
}

// DecodePhpBoot decode the PhpBoot object from request.
func DecodePhpBoot(req types.Request, decoder types.Decoder) (*appv1.PhpBoot, error) {
	bootType := req.AdmissionRequest.Kind.Kind

	var boot *appv1.PhpBoot
	if bootType == ApiTypePhp {
		boot = &v1.PhpBoot{}
		err := decoder.Decode(req, boot)
		if err != nil {
			return nil, err
		}
		return boot, nil
	}

	return boot, nil
}

// DecodePythonBoot decode the PythonBoot object from request.
func DecodePythonBoot(req types.Request, decoder types.Decoder) (*appv1.PythonBoot, error) {
	bootType := req.AdmissionRequest.Kind.Kind

	var boot *appv1.PythonBoot
	if bootType == ApiTypePython {
		boot = &v1.PythonBoot{}
		err := decoder.Decode(req, boot)
		if err != nil {
			return nil, err
		}
		return boot, nil
	}

	return boot, nil
}

// DecodeNodeJSBoot decode the NodeJSBoot object from request.
func DecodeNodeJSBoot(req types.Request, decoder types.Decoder) (*appv1.NodeJSBoot, error) {
	bootType := req.AdmissionRequest.Kind.Kind

	var boot *appv1.NodeJSBoot
	if bootType == ApiTypeNodeJS {
		boot = &v1.NodeJSBoot{}
		err := decoder.Decode(req, boot)
		if err != nil {
			return nil, err
		}
		return boot, nil
	}

	return boot, nil
}

// DecodeWebBoot decode the WebBoot object from request.
func DecodeWebBoot(req types.Request, decoder types.Decoder) (*appv1.WebBoot, error) {
	bootType := req.AdmissionRequest.Kind.Kind

	var boot *appv1.WebBoot
	if bootType == ApiTypeWeb {
		boot = &v1.WebBoot{}
		err := decoder.Decode(req, boot)
		if err != nil {
			return nil, err
		}
		return boot, nil
	}

	return boot, nil
}
