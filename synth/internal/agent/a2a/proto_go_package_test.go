package a2a

import (
	"reflect"
	"strings"
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"

	a2av1 "github.com/Mirelth/Mirelth-Synth/v7/synth/internal/agent/a2a/proto/a2a/v1"
)

// TestProtoGoPackageMatchesImportPath 守住 protobuf 描述符里的 go_package 与真实
// import 路径的一致性。
//
// 背景：模块路径变更时，go_package 被内嵌在 a2a.pb.go 的 rawDesc 常量里（裸字节），
// 前面还有两层长度前缀：FileOptions 负载长度、go_package 字符串长度。只改字符串而
// 不改长度（或多算少算一个字节、漏掉 go_package 的 tag 字节），编译能过但 init 阶段
// panic：slice bounds out of range [-N:]。本测试用反射取到的真实包路径反查描述符，
// 让这类漂移在测试阶段暴露，而不是线上崩溃。
func TestProtoGoPackageMatchesImportPath(t *testing.T) {
	pkgPath := reflect.TypeOf(a2av1.Task{}).PkgPath()
	if pkgPath == "" {
		t.Fatal("反射取不到 proto 包的 import 路径")
	}

	fd := a2av1.File_internal_agent_a2a_proto_a2a_v1_a2a_proto
	opts, ok := fd.Options().(*descriptorpb.FileOptions)
	if !ok {
		t.Fatalf("文件选项类型异常: %T", fd.Options())
	}
	goPackage := opts.GetGoPackage()
	if goPackage == "" {
		t.Fatal("描述符缺少 go_package")
	}

	// go_package 形如 "<import path>;<package alias>"
	decodePath := goPackage
	if i := strings.Index(goPackage, ";"); i >= 0 {
		decodePath = goPackage[:i]
	}
	if decodePath != pkgPath {
		t.Fatalf("描述符 go_package 与真实 import 路径不一致：\n  描述符: %s\n  实际  : %s\n"+
			"若刚改过模块路径，请同步 a2a.proto 的 go_package 与 a2a.pb.go rawDesc 的两层长度前缀",
			decodePath, pkgPath)
	}
}
