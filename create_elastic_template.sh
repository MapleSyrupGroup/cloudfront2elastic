#!/bin/bash

#Assumes your elastic instance is on localhost which it probably isn't.
curl -X POST "localhost:9200/_template/weblogs_1" -H 'Content-Type: application/json' -d'
{
  "index_patterns": ["weblogs*"],
  "mappings": {
    "properties":         {
      "date":               { "type": "date", "format": "yyyy-MM-dd HH:mm:ss" }, 
      "client_ip":          { "type": "ip" }, 
      "client_method":      { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "request_uri_stem":   { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "http_response_code": { "type": "integer"}, 
      "referer_page":       { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "client_user_agent":  { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "bot_detected":       { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "query_string":       { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "cf_edge_result":     { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "host_header":        { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "http_protocol":      { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "ssl_protocol":       { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}}, 
      "cf_client_result":   { "type": "text" ,"fields": { "raw": { "type":  "keyword" }}} 
    }
  }
}
'

