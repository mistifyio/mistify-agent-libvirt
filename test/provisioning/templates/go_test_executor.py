import os
import subprocess

workspace = os.environ['WORKSPACE'] if os.environ.get('WORKSPACE') else os.environ['HOME'] + '/gotest_workspace'
gopath = os.environ['GOPATH'] if os.environ.get('GOPATH') else os.environ['HOME'] + '/go'
repo = '{{mistify_agent_repo_name}}'
junit_report = workspace + '/test_results.xml'
temp_report = workspace + 'temp_results.txt'
whichgo = ''

try:
    whichgo = subprocess.check_output(['which', 'go'])
except subprocess.CalledProcessError as e:
    print "Go binary not found on path using default..."

gobinary = whichgo.strip() if whichgo else '/usr/local/go/bin/go'
gojunitbinary = gopath + '/bin/go-junit-report'

print "Using GOPATH:", gopath
print "Using go binary:", gobinary

print "Executing gotest..."
try:
    getcmd = [gobinary, 'get', '-t', './src/github.com/mistifyio/' + repo + '/...']
    print 'CMD:', ' '.join(getcmd)
    output = subprocess.check_output(getcmd, stderr=subprocess.STDOUT)
    print output
except subprocess.CalledProcessError as e:
    print "error>", e.output, '<'
    exit(1)

try:
    gotestcmd = ['sudo', '-E', gobinary, 'test', '-v', '-timeout', '30s', '-p', '1', './src/github.com/mistifyio/' + repo + '/...']
    print "CMD:", ' '.join(gotestcmd)

    temp_file = open(temp_report, "w")

    ps = subprocess.Popen((gotestcmd), stdout=subprocess.PIPE)

    lines_iterator = iter(ps.stdout.readline, b"")
    for line in lines_iterator:
        print(line)
        temp_file.write(line)

    ps.wait()
    temp_file.close()

    output = subprocess.check_output((gojunitbinary), stdin=open(temp_report,"r"))

    print "Writing test results to", junit_report
    file = open(junit_report, "w")
    file.write(output)
    file.close()
except subprocess.CalledProcessError as e:
    print "error>", e.output, '<'
    exit(1)

print "Finished... Exit(",ps.returncode,")"
exit(ps.returncode)