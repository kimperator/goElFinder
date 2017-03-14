package goElFinder

import (
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"io"
)

const APIver = "2.1"

/*
ElFinder connector handler

Example code:
	config := elFinder.Config{}
	config["l0"] = elFinder.Volume {
		Root: "./files/1",
		AllowDirs: []string{"/Allow"},
		DenyDirs:  []string{"/Deny"},
		DefaultRight: false,
	}
	config["l1"] = elFinder.Volume {
		Root: "./files/2",
		DefaultRight: true,
	}
	mux.Handle("/connector", elFinder.NetHttp(config))
 */

var conf Volumes

func NetHttp(config Volumes) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			volume response
			init, tree bool
			cmd string
			name string
			mode, bg string
			width, height, x, y, degree, quality int
			dirs []string
			id, path string
			paths map[string]string
			renames []string
			suffix string
			intersect []string
			chunk string
			uploadPath []string
			cid int
			err error
		)
		conf = config

		if r.Method == "GET" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println("GET:", r.Form)
			if r.Form["init"] != nil && r.Form["init"][0] == "1" {
				init = true
			} else {
				init = false
			}
			if r.Form["tree"] != nil && r.Form["tree"][0] == "1" {
				tree = true
			} else {
				tree = false
			}
			if r.Form["name"] != nil {
				name = r.Form["name"][0]
			}
			if r.Form["dirs[]"] != nil {
				dirs = r.Form["dirs[]"] // ToDo check rights
			}
			if r.Form["intersect[]"] != nil {
				intersect = r.Form["intersect[]"] // ToDo check rights
			}
			if r.Form["mode"] != nil {
				mode = r.Form["mode"][0]
			}
			if r.Form["width"] != nil {
				width, _  = strconv.Atoi(r.Form["width"][0])
			}
			if r.Form["height"] != nil {
				height, _  = strconv.Atoi(r.Form["height"][0])
			}
			if r.Form["x"] != nil {
				x, _  = strconv.Atoi(r.Form["x"][0])
			}
			if r.Form["y"] != nil {
				y, _  = strconv.Atoi(r.Form["y"][0])
			}
			if r.Form["degree"] != nil {
				degree, _  = strconv.Atoi(r.Form["degree"][0])
			}
			if r.Form["bg"] != nil {
				bg = r.Form["bg"][0]
			}
			/*if r.Form["quality"] != nil {
				quality, _  = strconv.Atoi(r.Form["quality"][0])
			}*/
			if r.Form["target"] != nil {
				id, path, err = parsePathHash(config, r.Form["target"][0])
				if err != nil {
					log.Println(err)
				}
				if !_getRight(id, path) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"error" : "errLocked"}`))
					return
				}

			} else if r.Form["targets[]"] != nil {
				paths = map[string]string{}
				for _, ft := range r.Form["targets[]"] {
					var i, p string
					i, p, err = parsePathHash(config, ft)
					if err != nil {
						log.Println(err)
					}
					if !_getRight(i, p) {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte(`{"error" : "errLocked"}`))
						return
					}
					paths[i] = p
				}
			} else {
				path = "/"
				return
			}

			if r.Form["cmd"] == nil {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error" : "errUnknownCmd"}`))
				return
			}
			cmd = r.Form["cmd"][0]
		} else if r.Method == "POST" {
			r.ParseMultipartForm(32 << 20) // ToDo check 8Mb
			fmt.Println("POST", r.PostForm)
			if r.PostForm["target"] != nil {
				id, path, err = parsePathHash(config, r.PostForm["target"][0])
				if err != nil {
					log.Println(err)
				}
				if !_getRight(id, path) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"error" : "errLocked"}`))
					return
				}

			} else if r.PostForm["targets[]"] != nil {
				paths = map[string]string{}
				for _, ft := range r.PostForm["targets[]"] {
					var i, p string
					i, p, err = parsePathHash(config, ft)
					if err != nil {
						log.Println(err)
					}
					if !_getRight(i, p) {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte(`{"error" : "errLocked"}`))
						return
					}
					paths[i] = p
				}
			}
			if r.PostForm["cid"] != nil {
				cid, err = strconv.Atoi(r.PostForm["cid"][0])
				if err != nil {
					log.Print(err)
				}
			}
			if r.PostForm["upload_path[]"] != nil {
				for u := range r.PostForm["upload_path[]"] {
					var i, p string
					i, p, err = parsePathHash(config, r.PostForm["upload_path[]"][u])
					if err != nil {
						log.Println(err)
					}
					if !_getRight(i, p) {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte(`{"error" : "errLocked"}`))
						return
					}
					uploadPath = append(uploadPath, p)
				}
			}
			if r.PostForm["renames[]"] != nil {
				renames = r.PostForm["renames[]"]
			}
			if r.PostForm["suffix"] != nil {
				suffix = r.PostForm["suffix"][0]
			}
			if r.PostForm["chunk"] != nil {
				chunk = r.PostForm["chunk"][0]
			}

			cmd = r.PostForm["cmd"][0]
		}
//-------------------------------------------------------------------------


		switch cmd {
		case "open":
			err := volume.open(id, path, init, tree)
			if err != nil {
				log.Println("Volume open:", err)
			}

		case "file":
			fileName, mimeType, data, err := volume.file(id, path)
			if err != nil {
				volume.Error = err.Error()
			} else {
				w.Header().Set("Content-Type", mimeType)
				if r.Form["download"] != nil {
					w.Header().Set("Content-Disposition", "attachment; filename='" + fileName + "'")
				} else {
					w.Header().Set("Content-Disposition", "inline; filename='" + fileName + "'")
				}
				w.Write(data)
				return
			}

		case "tree":
			err := volume.tree(id, path)
			if err != nil {
				volume.Error = err.Error()
			}
		case "parents":
			err := volume.parents(id, path)
			if err != nil {
				volume.Error = err.Error()
			}
		case "ls":
			volume.ls(id, path, intersect)
		case "tmb":
			err := volume.tmb(id, paths)
			if err != nil {
				volume.Error = err.Error()
			}
		case "size":
			//ToDo
		case "dim":
			err := volume.dim(id, path)
			if err != nil {
				volume.Error = err.Error()
			}
		case "mkdir":
			if len(dirs) > 0 {
				err := []string{}
				for _, f := range dirs {
					e := volume.mkdir(id, path, f)
					if e != nil {
						err = append(err, e.Error())
					}
				}
				if len(err) > 0 {
					volume.Error = err
				}
			} else {
				err = volume.mkdir(id, path, name)
				if err != nil {
					volume.Error = err.Error()
				}
			}

		case "mkfile":
			//ToDo
		case "rm":
			err := []string{}
			for i, f := range paths {
				e := volume.rm(i, f)
				if e != nil {
					err = append(err, e.Error())
				}
			}
			if len(err) > 0 {
				volume.Error = err
			}

		case "rename":
			err := volume.rename(id, path, name)
			if err != nil {
				volume.Error = err.Error()
			}

		case "duplicate":
			//ToDo
		case "paste":
			//ToDo
		case "upload":
			if r.PostForm["chunk"] != nil {
				var (

					file io.Reader
					err error
				)
				if r.PostForm["cid"] == nil {
					if len(renames) != 0 {
						fmt.Println("Result renames",volume.renames(id,path, suffix, renames))
					}
					fmt.Println("Result chunk merge", volume.chunkMerge(id, uploadPath[0], chunk))
				}
				for i := range r.MultipartForm.File["upload[]"] {
					file, err = r.MultipartForm.File["upload[]"][i].Open()
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println("Result chunk upload", volume.chunkUpload(cid, id, uploadPath[i], chunk, file))
				}

			} else {
				if len(renames) != 0 {
					fmt.Println("Result renames",volume.renames(id, path, suffix, renames))
				}
				esl := []string{}
				for i, f := range r.MultipartForm.File["upload[]"] {
					file, _ := f.Open()
					e := volume.upload(id, uploadPath[i], f.Filename, file)
					if e != nil {
						esl = append(esl, e.Error())
					}
					renames = []string{}
				}
				if len(esl) > 0 {
					volume.Error = esl
				}
			}


		case "get":
			//ToDo
		case "put":
			//ToDo
		case "archive":
			//ToDo
		case "extract":
			//ToDo
		case "search":
			//ToDo
		case "info":
			//ToDo
		case "resize":
			switch mode {
			case "resize":
				err = volume.resize(id, path, width, height)
			case "crop":
				err = volume.crop(id, path, x, y, width, height)
			case "rotate":
				err = volume.rotate(id, path, bg, degree)
			}
			_ = quality
			if err != nil {
				volume.Error = err.Error()
			}
		case "url":
		//	case "netmount":
		case "zipdl":
			//ToDo
		case "callback":
			//ToDo
		case "chmod":
			//ToDo


		default:
			volume.Error = "errUnknownCmd"

		}

		js, err := json.Marshal(volume)
		if err != nil {
			js = []byte(`{"error" : ["errConf", "errJSON"]}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	})
}
