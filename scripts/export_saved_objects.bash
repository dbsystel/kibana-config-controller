#!/bin/bash

progname=$0
VERBOSITY=0
LIST_ALL=0
EXPORT=0
EXPORT_ALL=0

SAVED_OBJECTS_TYPES="visualization dashboard search index-pattern config timelion-sheet"

: "${KIBANA_URL:=unknown}"
: "${KIBANA_SAVED_OBJECT_ID:=unknown}"
: "${KIBANA_SAVED_OBJECTS_DIRECTORY:=./saved_objects}"
: "${KIBANA_SAVED_OBJECTS_TYPE:=all}"
: "${KIBANA_LOGIN_PASSWORD_FILE:=unknown}"
: "${KIBANA_LOGIN_USER:=$(whoami)}"
: "${KIBANAi_NEW_INDEX_ID:=}"
: "${KIBANAi_OLD_INDEX_ID:=}"

function usage()
{
   cat >&2 << HEREDOC

   Usage: $progname ARGUMENTS

   required arguments:
   
     -k, --kibana-url URL                       specify the url of the corresponding kibana 
                                                (eg. https://my-kibana-url)
                                                
     -l, --list                                 specify if you wish to list all saved_objects in correspondig kibana
     OR                                         OR
     -e, --export                               if you wish to export one saved_object with given id (--id)
     
   
     -i, --id                                   specify saved_object id to export
     OR                                         OR
     -a, --all                                  if you wish to export all saved_objects from corresponding kibana
     
     
     -t, --type                                 type of saved object(is required if id is set)(visualization, dashboard, search, index-pattern, config, and timelion-sheet)

   optional arguments:
     -h, --help                                 show this help message and exit
     -v, --verbose                              increase the verbosity of the bash script
     -u, --user NAME                            specify kibana login username
     -p, --password-file PATH                   specify password file 
     -d, --directory PATH                       specify directory for exporting saved_objects
     -f, --filename NAME                        specify saved_object name in directory                                           (default: ./saved_objects/ (directory)

     -I, --index-id                             replace old-index-id with index-id in search
     -O, --old-index-id                         both paramether must be set in the same time

   examples:
        * list all saved_objects
          ./export_saved_objects.bash -k https://my-kibana-url -l
        * export one saved object with id 138e6750-2a0a-11e9-972a-999a33383a53 and type index-pattern
          ./export_saved_objects.bash -k https://my-kibana-url -e -i 138e6750-2a0a-11e9-972a-999a33383a53 -t index-pattern   
        * export one saved object with id 138e6750-2a0a-11e9-972a-999a33383a53 and type index-pattern into file mp.json
          ./export_saved_objects.bash -k https://my-kibana-url -e -i 138e6750-2a0a-11e9-972a-999a33383a53 -t index-pattern -f mp
        * export all saved_objects
          ./export_saved_objects.bash -k https://my-kibana-url -e -a
HEREDOC
}

log() {
	echo >&2 $*
}

export_one_saved_object() {
    local saved_object_saving_name=${KIBANA_SAVED_OBJECT_FILENAME:=`echo ${KIBANA_SAVED_OBJECTS_TYPE//_/-} | cut -c 1-32 | tr '[A-Z]' '[a-z]'`}
    saved_object_json=$(get_saved_object "$KIBANA_SAVED_OBJECT_ID")
    saved_object_title=$(echo "$saved_object_json" | jq -r .attributes.title | tr '[A-Z]' '[a-z]' | tr '_' '-' | tr '*' '-' )
    log "Downloading saved_object to: $KIBANA_SAVED_OBJECTS_DIRECTORY/$saved_object_saving_name-$saved_object_title.json"
    num_lines=$(echo "$saved_object_json" | wc -l);
    if [ "$num_lines" -le 6 ]; then
      log "ERROR:
  Couldn't retrieve saved_object $KIBANA_SAVED_OBJECT_ID! Maybe this saved_object does not exist!
      "
      exit 1
    fi
    if [ "$KIBANA_SAVED_OBJECTS_TYPE" == "search" ] && [ "$KIBANA_NEW_INDEX_ID" != "" ] && [ "$KIBANA_OLD_INDEX_ID" != "" ]; then
        saved_object_json=$(echo "$saved_object_json" | sed -e "s/$KIBANA_OLD_INDEX_ID/$KIBANA_NEW_INDEX_ID/g")
    fi
    echo "$saved_object_json" >$KIBANA_SAVED_OBJECTS_DIRECTORY/$saved_object_saving_name-$saved_object_title.json
}

export_all_saved_objects() {
 local saved_objects=$(list_saved_objects)
 local saved_object_json
  for saved_object in $saved_objects; do
    local saved_object_saving_name=`echo ${saved_object//_/-} | cut -c 1-32 | tr '[A-Z]' '[a-z]'`
    saved_object_json=$(get_saved_object "$saved_object")
    saved_object_title=$(echo "$saved_object_json" | jq -r .attributes.title | tr '[A-Z]' '[a-z]' | tr '_' '-' | tr '*' '-')
    log "Downloading saved_object to: $KIBANA_SAVED_OBJECTS_DIRECTORY/$KIBANA_SAVED_OBJECTS_TYPE-$saved_object_title.json"
    num_lines=$(echo "$saved_object_json" | wc -l);
    if [ "$num_lines" -le 6 ]; then
      log "ERROR:
  Couldn't retrieve saved_object $saved_object. Maybe this saved_object does not exist!
  Exit
      "
      exit 1
    fi
    if [ "$KIBANA_SAVED_OBJECTS_TYPE" == "search" ] && [ "$KIBANA_NEW_INDEX_ID" != "" ] && [ "$KIBANA_OLD_INDEX_ID" != "" ]; then
        saved_object_json=$(echo "$saved_object_json" | sed -e "s/$KIBANA_OLD_INDEX_ID/$KIBANA_NEW_INDEX_ID/g")
    fi
    echo "$saved_object_json" >$KIBANA_SAVED_OBJECTS_DIRECTORY/$KIBANA_SAVED_OBJECTS_TYPE-$saved_object_title.json
  done
}

get_saved_object() {
  local saved_object=$1

  if [[ -z "$saved_object" ]]; then
    log "ERROR:
  A saved_object must be specified.
  Exit
  "
    exit 1
  fi
 curl \
    --silent \
    --connect-timeout 10 --max-time 10 \
    -k \
    --user "$KIBANA_LOGIN_STRING" \
    $KIBANA_URL/api/saved_objects/$KIBANA_SAVED_OBJECTS_TYPE/$saved_object |
    jq 'del(.version, .updated_at)' 
}

list_saved_objects() {
  curl \
    --connect-timeout 10 --max-time 10 \
    --silent \
    -k \
    --user "$KIBANA_LOGIN_STRING" \
    $KIBANA_URL/api/saved_objects/_find?type=$KIBANA_SAVED_OBJECTS_TYPE |
    jq -r '.saved_objects[] | .id' |
    cut -d '/' -f2
}

get_saved_objects_list() {
  curl \
    --connect-timeout 10 --max-time 10 \
    --silent \
    -k \
    --user "$KIBANA_LOGIN_STRING" \
    $KIBANA_URL/api/saved_objects/_find?type=$KIBANA_SAVED_OBJECTS_TYPE |
    jq -r --arg ksot "$KIBANA_SAVED_OBJECTS_TYPE" '.saved_objects[] | "type: " + $ksot + " id: " + .id + " (" + .attributes.title + ")"' |
    cut -d '/' -f2 
}

function prepare() {
  log "Starting..."
  if [ "$KIBANA_LOGIN_PASSWORD_FILE" == "unknown" ]; then
      read -s -p "Please type in password for user $KIBANA_LOGIN_USER:" KIBANA_LOGIN_PASSWORD
      echo ""
      : "${KIBANA_LOGIN_STRING:=$KIBANA_LOGIN_USER:$KIBANA_LOGIN_PASSWORD}"
  else
    KIBANA_PASSWORD_FILE_CONTENT=`cat $KIBANA_LOGIN_PASSWORD_FILE`
    : "${KIBANA_LOGIN_STRING:=$KIBANA_LOGIN_USER:$KIBANA_LOGIN_PASSWORD_FILE_CONTENT}"
  fi


  [ -d $KIBANA_SAVED_OBJECTS_DIRECTORY ] || mkdir -p $KIBANA_SAVED_OBJECTS_DIRECTORY
}

function test_login() {
 log "Checking connection and authentication..."
 curl_response=$(curl --connect-timeout 10 --max-time 10 --write-out %{http_code} --silent --user "$KIBANA_LOGIN_STRING" --output /dev/null $KIBANA_URL/api/saved_objects/_find)
 if [ "$curl_response" -eq 200 ] ; then
   log "Authenticated - OK"
 else
    log "ERROR:
   Received http_code: $curl_response
   Exit
   "
   exit 1
 fi
}

function main() {
  prepare
  test_login
  if [ "$LIST_ALL" -gt 0 ]; then
	if [ "$KIBANA_SAVED_OBJECTS_TYPE" == "all" ]; then 
          log ""
          log "List of all saved_object of connected kibana:"
          log ""
          for saved_object_type in $SAVED_OBJECTS_TYPES; do
	    KIBANA_SAVED_OBJECTS_TYPE=$saved_object_type
	    get_saved_objects_list
	  done
	else 
          log ""
          log "List of all saved_object type $KIBANA_SAVED_OBJECTS_TYPE of connected kibana:"
          log ""
          get_saved_objects_list
	fi
  else
  	if [ "$EXPORT_ALL" -gt 0 ]; then
	  if [ "$KIBANA_SAVED_OBJECTS_TYPE" == "all" ]; then 
            for saved_object_type in $SAVED_OBJECTS_TYPES; do
              KIBANA_SAVED_OBJECTS_TYPE=$saved_object_type
	      export_all_saved_objects
	    done 
	  else 
                log "Starting export of all saved_objects type $KIBANA_SAVED_OBJECTS_TYPE to: $KIBANA_SAVED_OBJECTS_DIRECTORY"
  		export_all_saved_objects
	  fi
  	else
  		export_one_saved_object
  	fi
  fi
}

OPTS=$(getopt -o "k:lei:f:at:hvu:p:d:I:O:" --long "kibana-url:,list,export,id:,filename:,all,type:,help,verbose,user:,password-file:,directory:,index-id:,old-index-id" -n "$progname" -- "$@")
if [ $? -eq  0 ] ; then
  eval set -- "$OPTS"
  while true; do
    # uncomment the next line to see how shift is working
    # echo "\$1:\"$1\" \$2:\"$2\""
    case "$1" in
      -k | --kibana-url ) KIBANA_URL=$2; shift 2 ;;
      -l | --list ) LIST_ALL+=1; shift ;;
      -e | --export ) EXPORT+=1; shift ;;
      -i | --id ) KIBANA_SAVED_OBJECT_ID=$2; shift 2;;
      -f | --filename ) KIBANA_SAVED_OBJECT_FILENAME=$2; shift 2;;
      -a | --all ) EXPORT_ALL+=1; shift ;;
      -t | --type ) KIBANA_SAVED_OBJECTS_TYPE=$2; shift 2;;
      -h | --help ) usage; exit 0;;
      -v | --verbose ) VERBOSITY+=1; shift ;;
      -u | --user ) KIBANA_LOGIN_USER=$2; shift 2 ;;
      -p | --password-file ) KIBANA_LOGIN_PASSWORD_FILE=$2; shift 2 ;;
      -d | --directory ) KIBANA_SAVED_OBJECTS_DIRECTORY=$2; shift 2 ;;
      -I | --index-id ) KIBANA_NEW_INDEX_ID=$2; shift 2;;
      -O | --old-index-id ) KIBANA_OLD_INDEX_ID=$2; shift 2;;
      -- ) shift; break ;;
      * ) break ;;
    esac
  done
  
  if [ "$KIBANA_URL" == "unknown" ] ||
     ( [ $LIST_ALL -eq 0 ] && [ $EXPORT -eq 0 ] ) ||
     ( [ $LIST_ALL -gt 0 ] && [ $EXPORT -gt 0 ] ) ||
     ( [ $LIST_ALL -eq 0 ] && [ $EXPORT_ALL -eq 0 ] && [ "$KIBANA_SAVED_OBJECT_ID"  == "unknown" ] ) ||
     ( [ $LIST_ALL -eq 0 ] && [ $EXPORT_ALL -eq 0 ] && [ "$KIBANA_SAVED_OBJECT_ID"  != "unknown" ] && [ "$KIBANA_SAVED_OBJECTS_TYPE"  == "all" ]) ||
     ( [ "$KIBANA_NEW_INDEX_ID" == "" ] && [ "$KIBANA_NEW_INDEX_ID" != "" ] ) ||
     ( [ "$KIBANA_NEW_INDEX_ID" != "" ] && [ "$KIBANA_NEW_INDEX_ID" == "" ] ) ;then

     usage
     exit 1
  fi 
  if [ $VERBOSITY -gt 0 ]; then

     cat << DEBUG_OUTPUT

     Debug Output:

     KIBANA_URL:                         $KIBANA_URL
     KIBANA_LOGIN_USER:                  $KIBANA_LOGIN_USER
     KIBANA_LOGIN_PASSWORD_FILE          $KIBANA_LOGIN_PASSWORD_FILE
     KIBANA_SAVED_OBJECTS_DIRECTORY:        $KIBANA_SAVED_OBJECTS_DIRECTORY
     KIBANA_SAVED_OBJECT_ID:              $KIBANA_SAVED_OBJECT_ID

DEBUG_OUTPUT
     fi

  main
else
  log "Error in command line arguments." >&2
  usage
fi