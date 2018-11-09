#!/usr/bin/python3
import subprocess
import os
import re
import json
import argparse

def getPort(): 
    portre = re.compile(r'port := \"([\:\d]+)\"')
    #look for a server.go, if none, get a main.go, if none, return ""
    for f in os.listdir(os.getcwd()):
        if f == 'server.go' or f == 'main.go':
            fi = open(os.path.join(os.getcwd(), f))
            txt = fi.read()
            fi.close()

            res = portre.findall(txt)
            print(res)
            if len(res) > 0: 
                return res[0] 
    return ""

parser = argparse.ArgumentParser(description='configure the config.json file')
parser.add_argument("-l", dest='loc', action='store_const', const=True, help='Print out the location of the variables')
parser.add_argument("-v", dest='verbose', action='store_const', const=True, help='verbose printing')
args = parser.parse_args()

rtext = r'(((.*\n){1})(.*os\.Getenv\("([a-z,A-Z,_,0-9]+)"\)))'
ctext = r'.*(//).*os\.Getenv\(\"([a-z,A-Z,_,0-9]+?)\"\)'

v = re.compile(rtext)
commentre = re.compile(ctext)

#filesToFind = subprocess.run(r"""go list""", stdout=subprocess.PIPE, shell=True)
filesToFind = subprocess.run(r""" go list -f '{{ join .Deps "\n" }}' """, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True).stdout.decode("utf-8")
toCheck = list(filter(lambda s: ('byuoitav' in s and 'vendor' not in s), filesToFind.split('\n')))

envVars = set()

#now we have to go to each of those directories - assume a GOPATH for now
base = os.environ["GOPATH"] + "/src/"
toCheck.append(os.getcwd().split("/src/")[1])
for i in toCheck:
    for f in os.listdir(base+i):
        if not os.path.isfile(os.path.join(base, i , f)):
            continue

        if args.verbose:
            print(os.path.join(i,f))

        try:
            fi = open(os.path.join(base, i, f))
            txt = fi.read()
            fi.close()
        except:
            continue

        res = v.findall(txt)
        for r in res:
            #check for the DEPLOY_NOT_REQUIRED flat on the line before
            if "+deploy not_required" not in r[2]:
                res = commentre.findall(r[3])
                if len(res) is not 0 :
                    if args.verbose:
                        print(r[4]," - comment, skipping")
                    continue

                if args.loc:
                    print(r[4] , " - " , os.path.join(i,f))
                envVars.add(r[4])
        

print(envVars)
configPath = os.path.join(os.getcwd(), "config.json")

if os.path.exists(configPath): #open the config file
    try: 
        fi = open(configPath, 'r')
        data = json.loads(fi.read())
        fi.close()
        data["env-vars"] = sorted(list(envVars))
    except:
        data = {
                "name": os.getcwd().split("/")[-1],
                "port": getPort(),
                "env-vars": sorted(list(envVars))
                }
        
    fi = open(configPath, 'w')
    fi.write(json.dumps(data, indent=4))
    fi.close()

else: #We need to build it
    data = {
            "name": os.getcwd().split("/")[-1],
            "port": getPort(),
            "env-vars": sorted(list(envVars))
            }
    fi = open(configPath, 'w')
    fi.write(json.dumps(data, indent=4))
    fi.close()

