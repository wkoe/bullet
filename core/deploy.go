package core

import (
	"fmt"
	"log"
	"os"

	"github.com/FurqanSoftware/bullet/distro"
	_ "github.com/FurqanSoftware/bullet/distro/ubuntu"
	"github.com/FurqanSoftware/bullet/spec"
	"github.com/FurqanSoftware/bullet/ssh"
)

func Deploy(nodes []Node, spec *spec.Spec, rel *Release) error {
	for _, n := range nodes {
		log.Printf("Connecting to %s", n.Addr())
		c, err := ssh.Dial(n.Addr(), n.Identity)
		if err != nil {
			return err
		}

		d, err := distro.New(c)
		if err != nil {
			return err
		}

		err = deployNode(n, c, d, spec, rel)
		if err != nil {
			return err
		}
	}
	return nil
}

func deployNode(n Node, c *ssh.Client, d distro.Distro, spec *spec.Spec, rel *Release) error {
	log.Print("Uploading tarball")
	tarPath := fmt.Sprintf("/tmp/%s-%s.tar.gz", spec.Application.Identifier, rel.Hash)
	err := uploadTarball(c, tarPath, rel.Tarball)
	if err != nil {
		return err
	}

	log.Print("Extracting tarball")
	relDir := fmt.Sprintf("/opt/%s/releases/%s", spec.Application.Identifier, rel.Hash)
	err = d.ExtractTar(tarPath, relDir)
	if err != nil {
		return err
	}
	log.Print("Removing tarball")
	err = d.Remove(tarPath)
	if err != nil {
		return err
	}

	log.Print("Updating current marker")
	err = d.Symlink(relDir, fmt.Sprintf("/opt/%s/current", spec.Application.Identifier))
	if err != nil {
		return err
	}

	log.Print("Building images")
	for _, p := range spec.Application.Programs {
		err = d.Build(spec.Application, p)
		if err != nil {
			return err
		}
	}

	log.Print("Restarting containers")
	for _, k := range spec.Application.ProgramKeys {
		p := spec.Application.Programs[k]
		err = d.RestartAll(spec.Application, p)
		if err != nil {
			return err
		}
	}

	log.Print("Removing stale releases")
	err = d.Prune(fmt.Sprintf("/opt/%s/releases", spec.Application.Identifier), 5)
	if err != nil {
		return err
	}

	return nil
}

func uploadTarball(c *ssh.Client, dst string, tar Tarball) error {
	f, err := os.Open(tar.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return err
	}
	return c.Push(dst, s.Mode(), s.Size(), f)
}
