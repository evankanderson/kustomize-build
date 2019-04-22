package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var updateDir *string = flag.String("configs", ".", "What directory to search for configurations.")

// CollectObjs returns ObjectReferences for all the recognized Kubernetes
// objects (YAML objects with recognized TypeMeta and name) in a specified
// file.
func CollectObjs(path string) ([]v1.ObjectReference, error) {
	objRefs := make([]v1.ObjectReference, 0)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var u unstructured.Unstructured
	dec := yaml.NewYAMLOrJSONDecoder(f, 4096)
	for true {
		err = dec.Decode(&u) 
		if err != nil {
			break
		}
		objRefs = append(objRefs, v1.ObjectReference{
			APIVersion: u.GetAPIVersion(),
			Kind:       u.GetKind(),
			Namespace:  u.GetNamespace(),
			Name:       u.GetName(),
		})
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return objRefs, nil
}

func main() {
	flag.Parse()
	files := make([]string, 0)
	objRefs := make([]v1.ObjectReference, 0)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Unable to read %s", path)
			return nil
		}
		if info.IsDir() && path != "." && strings.HasPrefix(filepath.Base(path), ".") {
			return filepath.SkipDir
		}
		if info.Mode().IsRegular() {
			files = append(files, path)
			objs, err := CollectObjs(path)
			if err == nil {
				objRefs = append(objRefs, objs...)
			}
		}
		return nil
	}

	err := filepath.Walk(*updateDir, walker)
	if err != nil {
		log.Fatalf("Error walking %s: %v", *updateDir, err)
	}

	log.Printf("Found the following in %q: %s", *updateDir, files)
	log.Printf("Objects: %s", objRefs)
}
