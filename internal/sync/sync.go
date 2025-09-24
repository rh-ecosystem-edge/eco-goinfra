package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

type repo struct {
	Sync               bool                `yaml:"sync"`
	Name               string              `yaml:"name"`
	RepoLink           string              `yaml:"repo_link"`
	Branch             string              `yaml:"branch"`
	RemoteAPIDirectory string              `yaml:"remote_api_directory"`
	LocalAPIDirectory  string              `yaml:"local_api_directory"`
	ReplaceImports     []map[string]string `yaml:"replace_imports"`
	Excludes           []string            `yaml:"excludes"`
}

func main() {
	// Enable glog output
	_ = flag.Lookup("logtostderr").Value.Set("true")
	_ = flag.Lookup("v").Value.Set("100")
	configFiles := flag.String("config-file", "internal/sync/configs", "path to config files")
	flag.Parse()

	klog.V(100).Info("Loading config file")

	config := newConfig(*configFiles)

	klog.V(100).Info("Initiating repository sync")

	for _, repo := range config {
		if repo.Sync {
			klog.V(100).Infof("#### Syncing repo %s ####", repo.Name)
			syncRemoteRepo(&repo)
		} else {
			klog.V(100).Infof("Sync disabled for repo %s. Skip", repo.Name)
		}
	}
}

func syncRemoteRepo(repo *repo) {
	klog.V(100).Infof("Syncing repo: %s, destination repo link: %s", repo.Name, repo.RemoteAPIDirectory)

	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	projectClonedDirectory := path.Join(basePath, repo.Name, repo.RemoteAPIDirectory)
	projectLocalDirectory := path.Join("./pkg", repo.LocalAPIDirectory)

	gitClone(basePath, repo)
	excludeAndRefactor(projectClonedDirectory, projectLocalDirectory, repo)

	klog.V(100).Infof("Comparing local %s and cloned %s api directories for repo %s",
		projectLocalDirectory, projectClonedDirectory, repo.Name)

	err := execCmd("", "diff", []string{projectClonedDirectory, projectLocalDirectory})
	if err != nil {
		klog.V(100).Infof("Repos not synced. Copying cloned repo %s to %s", projectClonedDirectory, projectLocalDirectory)

		copyClonedToLocal(projectClonedDirectory, projectLocalDirectory)
	}

	klog.V(100).Infof("Remove cloned directory from filesystem: %s", path.Join(basePath, repo.Name))

	if _, err := os.Stat(path.Join(basePath, repo.Name)); !os.IsNotExist(err) {
		err := os.RemoveAll(path.Join(basePath, repo.Name))
		if err != nil {
			klog.V(100).Infof("Failed to remove cloned directory %s exit with error code 1",
				path.Join(basePath, repo.Name))
			os.Exit(1)
		}
	}
}

// excludeAndRefactor excludes and refactors files in the clonedDir to prepare them for being compared or copied to the
// localDir.
func excludeAndRefactor(clonedDir, localDir string, repo *repo) {
	klog.V(100).Infof("Updating %s to match expected state of %s", clonedDir, localDir)

	if len(repo.Excludes) > 0 {
		klog.V(100).Infof("Remove excluded files under %s", path.Base(clonedDir))

		err := excludeFiles(clonedDir, repo.Excludes...)
		if err != nil {
			klog.V(100).Infof("Failed to remove excluded files due to %v. Exit with error 1", err)
			os.Exit(1)
		}
	}

	klog.V(100).Infof("Replace cloned package name: %s with the local package name: %s",
		path.Base(clonedDir), path.Base(localDir))

	err := refactor(
		fmt.Sprintf("package %s", path.Base(clonedDir)),
		fmt.Sprintf("package %s", path.Base(localDir)),
		clonedDir, "*.go")
	if err != nil {
		klog.V(100).Infof("Failed to replace package names due to %v. Exit with error 1", err)
		os.Exit(1)
	}

	for _, importMap := range repo.ReplaceImports {
		err = refactor(importMap["old"], importMap["new"], clonedDir, "*.go")
		if err != nil {
			klog.V(100).Infof("Failed to refactor files due to %v. Exit with error 1", err)
			os.Exit(1)
		}
	}
}

func copyClonedToLocal(clonedDir, localDir string) {
	klog.V(100).Infof("Create path to new local directory: %s", localDir)

	// We use MkdirAll to make sure the path leading up to localDir exists.
	if os.MkdirAll(localDir, 0750) != nil {
		klog.V(100).Info("Failed to create local directory. Exit with error code 1")
		os.Exit(1)
	}

	// We use RemoveAll to delete just localDir but not the path leading to it.
	err := os.RemoveAll(localDir)
	if err != nil {
		klog.V(100).Infof("Failed to remove old local directory %s due to %v. Exit with error 1", localDir, err)
		os.Exit(1)
	}

	err = execCmd("", "cp", []string{"-a", clonedDir, localDir})
	if err != nil {
		klog.V(100).Infof("Failed to sync directories due to %v. Exit with error 1", err)
		os.Exit(1)
	}
}

func gitClone(localPath string, repo *repo) {
	klog.V(100).Infof("Cloning repo %s from %s", repo.Name, repo.RepoLink)
	localDirectory := path.Join(localPath, repo.Name)

	if _, err := os.Stat(localDirectory); !os.IsNotExist(err) {
		klog.V(100).Infof(
			"Local directory already exists for the project: %s. Removing directory", repo.Name)

		err := os.RemoveAll(localDirectory)
		if err != nil {
			klog.V(100).Infof("Failed to remove repo directory %s due to %v exit with error code 1",
				localDirectory, err)
			os.Exit(1)
		}
	}

	err := execCmd(
		localPath,
		"git",
		[]string{"clone", "-n", "--depth=1", "--filter=tree:0", "-b", repo.Branch, repo.RepoLink, repo.Name})
	if err != nil {
		klog.V(100).Info("Failed to clone repo due to cmd error. Exit with error code 1")
		os.Exit(1)
	}

	err = execCmd(localDirectory, "git", []string{"sparse-checkout", "set", "--no-cone", repo.RemoteAPIDirectory})
	if err != nil {
		klog.V(100).Info("Failed to sparse-checkout repo due to cmd error. Exit with error code 1")
		os.Exit(1)
	}

	err = execCmd(localDirectory, "git", []string{"checkout"})
	if err != nil {
		klog.V(100).Info("Failed to checkout repo due to cmd error. Exit with error code 1")
		os.Exit(1)
	}
}

func execCmd(dirName, binary string, args []string) error {
	klog.V(100).Infof("Executing cmd: %s, with args: %v, in directory: %s", binary, args, dirName)

	cmd := exec.Command(binary, args...)

	if dirName != "" {
		cmd.Dir = dirName
	}

	out, err := cmd.Output()
	if err != nil {
		klog.V(100).Infof("Failed to execute cmd due to %s. Output: %s", err, string(out))

		return err
	}

	return nil
}

func newConfig(pathToConfigFiles string) []repo {
	var repoConfigs []repo

	klog.V(100).Info("Read files in configs directory")

	err := filepath.Walk(pathToConfigFiles, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			klog.V(100).Infof("Loading config file %s", info.Name())

			var config []repo

			err := readFile(&config, path)
			if err != nil {
				klog.V(100).Infof("Error to read config file: %v", err)

				return err
			}

			repoConfigs = append(repoConfigs, config...)
		}

		return nil
	})
	if err != nil {
		klog.V(100).Infof("Error to list files in directory %s due to %v", pathToConfigFiles, err)

		return nil
	}

	if len(repoConfigs) == 0 {
		klog.V(100).Info("Config directory is empty")

		return nil
	}

	return repoConfigs
}

func readFile(cfg *[]repo, cfgFile string) error {
	openedCfgFile, err := os.Open(cfgFile)
	if err != nil {
		return err
	}

	defer openedCfgFile.Close()

	decoder := yaml.NewDecoder(openedCfgFile)

	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func refactor(oldLine, newLine, root string, patterns ...string) error {
	return filepath.WalkDir(root, refactorFunc(oldLine, newLine, patterns))
}

func refactorFunc(oldLine, newLine string, filePatterns []string) fs.WalkDirFunc {
	return func(filePath string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirEntry.IsDir() {
			return nil
		}

		var matched bool

		for _, pattern := range filePatterns {
			var err error

			matched, err = filepath.Match(pattern, dirEntry.Name())
			if err != nil {
				return err
			}

			if matched {
				oldContents, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}

				klog.V(100).Infof("Refactoring: %s", filePath)

				newContents := strings.ReplaceAll(string(oldContents), oldLine, newLine)

				err = os.WriteFile(filePath, []byte(newContents), 0)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func excludeFiles(path string, patterns ...string) error {
	return filepath.WalkDir(path, excludeFunc(path, patterns))
}

func excludeFunc(root string, patterns []string) fs.WalkDirFunc {
	return func(filePath string, dirEntry fs.DirEntry, err error) error {
		if filePath == root {
			return nil
		}

		for _, pattern := range patterns {
			match, err := filepath.Match(pattern, filepath.Base(filePath))
			if err != nil {
				return err
			}

			if match {
				klog.V(100).Infof("Found that path %s matches %s - Excluding", filePath, pattern)

				return os.RemoveAll(filePath)
			}
		}

		return nil
	}
}
