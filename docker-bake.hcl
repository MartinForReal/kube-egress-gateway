
group "default" {
  targets = ["daemon", "controller", "cnimanager", "cni", "cni-ipam"]
}

target "base" {
  dockerfile = "base.Dockerfile"
}

target "daemon-compile" {
  inherits = ["base"]
  args = {
    MAIN_ENTRY = "kube-egress-gateway-daemon",
    BASE_IMAGE = "mcr.microsoft.com/aks/devinfra/base-os-runtime-nettools:master.221105.1",
  }
}
target "daemon" {
  inherits = ["daemon-tags"]
  dockerfile = "root.Dockerfile"
  contexts = {
    baseimg = "target:daemon-compile"
  }
}

target "controller" {
  inherits = ["base","controller-tags"]
  args = {
    MAIN_ENTRY = "kube-egress-gateway-controller",
  }
}

target "cnimanager-compile" {
  inherits = ["base"]
  args = {
    MAIN_ENTRY = "kube-egress-gateway-cnimanager",
  }
}

target "cnimanager" {
  inherits = ["cnimanager-tags"]
  dockerfile = "cnimanager.Dockerfile"
  contexts = {
    baseimg = "target:cnimanager-compile"
  }
  args = {
    GRPC_HEALTH_PROBE_VERSION = "v0.4.14"
  }
}
target "cni-compile" {
  inherits = ["base"]
  args = {
    MAIN_ENTRY = "kube-egress-cni",
  }
}
target "cni" {
  inherits = ["cni-tags"]
  dockerfile = "cni.Dockerfile"
  contexts = {
    baseimg = "target:cni-compile"
  }
}

target "cni-ipam-compile" {
  inherits = ["base"]
  args = {
    MAIN_ENTRY = "kube-egress-cni-ipam",
  }
}
target "cni-ipam" {
  inherits = ["cni-ipam-tags"]
  dockerfile = "cni.Dockerfile"
  contexts = {
    baseimg = "target:cni-ipam-compile"
  }
}