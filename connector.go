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
func NetHttp(config Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			volume response
			init, tree bool
			cmd string
			name string
			dirs []string
			target string
			targets []string
			renames []string
			suffix string
			intersect []string
			chunk string
			uploadPath []string
			cid int
			err error
		)

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
			if r.Form["target"] != nil {
				volume, target, err = parseHash(config, r.Form["target"][0])
				if err != nil {
					log.Println(err)
				}
				if !volume.checkRight(target) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"error" : "errLocked"}`))
					return
				}

			} else if r.Form["targets[]"] != nil {
				for _, ft := range r.Form["targets[]"] {
					var p string
					volume, p, err = parseHash(config, ft)
					if err != nil {
						log.Println(err)
					}
					if !volume.checkRight(p) {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte(`{"error" : "errLocked"}`))
						return
					}
					targets = append(targets, p)
				}
			} else {
				target = "/"
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
				volume, target, err = parseHash(config, r.PostForm["target"][0])
				if err != nil {
					log.Println(err)
				}
				if !volume.checkRight(target) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"error" : "errLocked"}`))
					return
				}

			} else if r.PostForm["targets[]"] != nil {
				for _, ft := range r.PostForm["targets[]"] {
					var p string
					volume, p, err = parseHash(config, ft)
					if err != nil {
						log.Println(err)
					}
					//ToDo error multi path
					if !volume.checkRight(p) {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte(`{"error" : "errLocked"}`))
						return
					}
					targets = append(targets, p)
				}
			}
			if r.PostForm["cid"] != nil {
				cid, err = strconv.Atoi(r.PostForm["cid"][0])
				if err != nil {
					log.Print(err)
				}
			}
			if r.PostForm["upload_path[]"] != nil {
				for i := range r.PostForm["upload_path[]"] {
					var p string
					volume, p, err = parseHash(config, r.PostForm["upload_path[]"][i])
					if err != nil {
						log.Println(err)
					}
					//ToDo error multi path
					if !volume.checkRight(p) {
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
			err := volume.open(init, tree, target)
			if err != nil {
				log.Println(err)
			}

		case "file":
			fileName, mimeType, data, err := volume.file(target)
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
		case "parents":
		case "ls":
			volume.ls(target, intersect)
		case "tmb":
		case "size":

		case "dim":
			err := volume.dim(target)
			if err != nil {
				volume.Error = err.Error()
			}

		case "mkdir":
			if len(dirs) > 0 {
				err := []string{}
				for _, f := range dirs {
					e := volume.mkdir(target, f)
					if e != nil {
						err = append(err, e.Error())
					}
				}
				if len(err) > 0 {
					volume.Error = err
				}
			} else {
				err = volume.mkdir(target, name)
				if err != nil {
					volume.Error = err.Error()
				}
			}

		case "mkfile":
		case "rm":
			err := []string{}
			for _, f := range targets {
				e := volume.rm(f)
				if e != nil {
					err = append(err, e.Error())
				}
			}
			if len(err) > 0 {
				volume.Error = err
			}

		case "rename":
			err := volume.rename(target, name)
			if err != nil {
				volume.Error = err.Error()
			}

		case "duplicate":
		case "paste":
		case "upload":
			if r.PostForm["chunk"] != nil {
				var (

					file io.Reader
					err error
				)
				if r.PostForm["cid"] == nil {
					if len(renames) != 0 {
						fmt.Println("Result renames",volume.renames(target, suffix, renames))
					}
					fmt.Println("Result chunk merge", volume.chunkMerge(target, chunk))
				}
				for i := range r.MultipartForm.File["upload[]"] {
					file, err = r.MultipartForm.File["upload[]"][i].Open()
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println("Result chunk upload", volume.chunkUpload(cid, uploadPath[i], chunk, file))
				}

			} else {
				if len(renames) != 0 {
					fmt.Println("Result renames",volume.renames(target, suffix, renames))
				}
				esl := []string{}
				for i, f := range r.MultipartForm.File["upload[]"] {
					file, _ := f.Open()
					e := volume.upload(uploadPath[i], f.Filename, file)
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
		case "put":
		case "archive":
		case "extract":
		case "search":
		case "info":
		case "resize":
		case "url":
		//	case "netmount":
		case "zipdl":
		case "callback":
		case "chmod":


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
