package observer

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/suite"
	"testing"
)

var (
	FilterRaw = `{
		"type": "(patchset-created|ref-updated)",
		"patchSet": {
		  "uploader": {
			"email": "guoxuxing@wandoujia.com"
		  }
		},
		"change": {
		  "branch": "release-.+",
		  "project": "loki"
		}
	  }`
	RefUpdateRaw = `{
	   "refUpdate" : {
		  "refName" : "refs/tags/update_uni_crawler_master_staging_success",
		  "oldRev" : "2fe23dd68dd6460ba88fa80a0a1129290ba7bedd",
		  "newRev" : "deec55281d04a64d2483deabaff436a284ce60a7",
		  "project" : "orion/crawler"
	   },
	   "type" : "ref-updated",
	   "submitter" : {
		  "email" : "tanyi@wandoujia.com",
		  "username" : "tanyi",
		  "name" : "Yi TAN"
	   }
	}`
	PatchSetCreatedRaw = `{
   "type" : "patchset-created",
   "uploader" : {
      "email" : "guoxuxing@wandoujia.com",
      "username" : "zengyaopeng",
      "name" : "Yaopeng Zeng"
   },
   "patchSet" : {
      "author" : {
         "email" : "zengyaopeng@wandoujia.com",
         "username" : "zengyaopeng",
         "name" : "Yaopeng Zeng"
      },
      "parents" : [
         "2fe23dd68dd6460ba88fa80a0a1129290ba7bedd"
      ],
      "sizeInsertions" : 347,
      "revision" : "deec55281d04a64d2483deabaff436a284ce60a7",
      "isDraft" : false,
      "createdOn" : 1447676592,
      "sizeDeletions" : -30,
      "number" : "1",
      "uploader" : {
         "username" : "zengyaopeng",
         "name" : "Yaopeng Zeng",
         "email" : "guoxuxing@wandoujia.com"
      },
      "ref" : "refs/changes/16/121516/1"
   },
   "change" : {
      "subject" : "修复上周的一些bug",
      "commitMessage" : "修复上周的一些bughttps://app.asana.com/0/12109140235743/64371773476756",
      "status" : "NEW",
      "project" : "loki",
      "owner" : {
         "email" : "zengyaopeng@wandoujia.com",
         "name" : "Yaopeng Zeng",
         "username" : "zengyaopeng"
      },
      "url" : "http://git.wandoulabs.com/121516",
      "number" : "121516",
      "branch" : "release-2015",
      "id" : "Ideec55281d04a64d2483deabaff436a284ce60a7"
   }
}`
)

func TestMsgCompare(t *testing.T) {
	filter := make(map[string]interface{})
	compiledFilter := make(map[string]interface{})
	msg := make(map[string]interface{})
	err := json.Unmarshal([]byte(FilterRaw), &filter)
	if !assert.Nil(t, err) {
		t.Log(err)
	}
	err = mustCompileFilter(filter, compiledFilter)
	if !assert.Nil(t, err) {
		t.Log(err)
	}
	err = json.Unmarshal([]byte(RefUpdateRaw), &msg)
	if !assert.Nil(t, err) {
		t.Log(err)
	}
	ret, _ := msgCompare(compiledFilter, msg)
	assert.False(t, ret)
	err = json.Unmarshal([]byte(PatchSetCreatedRaw), &msg)
	assert.Nil(t, err)
	ret, _ = msgCompare(compiledFilter, msg)
	assert.True(t, ret)
}
