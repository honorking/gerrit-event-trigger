#!/bin/env python
#-*- coding: utf-8 -*-

import requests

url = "http://gerrit-observatory.internal.wandoujia.com/observers"

def main():
    data = {
      "filter": {
        "type": "ref-updated",
        "change": {
          "branch": "release*",
        }
      },
      "hook_url": "http://docker.internal.wandoujia.com/image/gerrit", 
      "comment": "构建 docker 镜像"
    }
    
    ret = requests.post(url, json=data)
    print ret.status_code


if __name__ == "__main__":
    main()
