#! /bin/bash

set -eu

flagFile=/var/run/drobomon/droboLastGoodReport
firstFailure=/var/run/drobomon/droboFirstFailure
logDir=/var/log/drobomon
configFile=/home/ajm/.config/drobomon/env

. $configFile

healthLog=${logDir}/health.log
statusLog=${logDir}/status.log
iftttKey=$DROBOMON_KEY

# Only report success once in a while to prevent excessive load.
goodReportFreq=$(( 86400 * 7 ))

if [[ ! -d $(dirname $flagFile) ]]; then
	mkdir -p $(dirname $flagFile)
fi

if [[ ! -d $(dirname $logDir) ]]; then
	mkdir -p $(dirname $logDir)
fi

if [[ x"$logDir" == "x" ]]; then
    printf "No log dir!\n"
    exit 1
fi

if [[ x"$iftttKey" == "x" ]]; then
    printf "No IFTTT Key!\n"
    exit 1
fi


printf "Drobomon:$(date -Iseconds):" | tee -a ${statusLog}
curl -s 192.168.2.53:5000/v1/drobomon/status --output - | tee -a ${statusLog}

health=$(curl -s 192.168.2.53:5000/v1/drobomon/health --output -)

if [ $? -ne 0 ]; then
    printf "Drobomon:$(date -Iseconds):Fail due to curl status\n" | tee -a ${healthLog}
    curl -X POST https://maker.ifttt.com/trigger/drobo_status_fail/with/key/${iftttKey}
    exit 0
fi

healthString=$(echo $health | jq -r .status)

if [[ "$healthString" == "fail" ]]; then

    if [[ -f $firstFailure ]]; then
    	printf "Drobomon:$(date -Iseconds):Fail due to drobo status - with notification\n" | tee -a ${healthLog}
	rm -f $firstFailure
        curl -X POST https://maker.ifttt.com/trigger/drobo_status_fail/with/key/${iftttKey}
    else
	printf "Drobomon:$(date -Iseconds):First Fail due to drobo status - defer notification\n" | tee -a ${healthLog}
	touch $firstFailure
    fi
        exit 0
fi


if [[ "$healthString" == "warn" ]]; then
    printf "Drobomon:$(date -Iseconds):Warn due to drobo status\n" | tee -a ${healthLog}
    curl -X POST https://maker.ifttt.com/trigger/drobo_status_warn/with/key/${iftttKey}
    exit 0
fi

if [[ -f $flagFile ]]; then
    secondsSinceLastGood=$(( $(date +%s) - $(stat -c '%Y' $flagFile) ))
else
    secondsSinceLastGood=$(( $goodReportFreq + 1 ))
fi

if [[ "$healthString" == "pass"  ]]; then
    if [[ -f $firstFailure ]]; then
        rm $firstFailure
    fi

    if [[ $secondsSinceLastGood -gt $goodReportFreq ]]; then
	printf "Drobomon:$(date -Iseconds):Pass - with notification\n" | tee -a ${healthLog}
        curl -X POST https://maker.ifttt.com/trigger/drobo_status_pass/with/key/${iftttKey}
        touch $flagFile
    else
	printf "Drobomon:$(date -Iseconds):Pass - no notification\n" | tee -a ${healthLog}
    fi
    exit 0
fi

printf "Drobomon:$(date -Iseconds):ERROR - shouldn't get here!\n" | tee -a ${healthLog}
exit 1

