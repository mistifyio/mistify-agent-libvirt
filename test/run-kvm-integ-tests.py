import git, os, shutil, subprocess

repo_name = 'infrastructure-jenkins-slave'
repo_remote = "git@github.com:mistifyio/" + repo_name
branch = 'master'
checkout_dir = '/tmp/' + repo_name
root_dir = os.path.dirname(os.path.realpath(__file__))
vars_dir = root_dir + '/provisioning/vars/'
if os.path.isdir(checkout_dir):
    shutil.rmtree(checkout_dir)

repo = git.Repo.clone_from(repo_remote, checkout_dir, branch=branch)
container_provisioning_roles_dir = root_dir + '/provisioning/roles/'
infrastructure_roles = os.listdir(checkout_dir + '/roles/')

for role in infrastructure_roles:
    if os.path.isdir(container_provisioning_roles_dir + role):
    	shutil.rmtree(container_provisioning_roles_dir + role)
    print "Copying role", role, "from", repo_name
    shutil.copytree(checkout_dir + '/roles/' + role, container_provisioning_roles_dir + role)

if not os.path.exists(vars_dir):
    os.makedirs(vars_dir)

print "Copying vars files from",checkout_dir + '/vars/vaulted_vars', "from", repo_name
shutil.copy(checkout_dir + '/vars/vaulted_vars', vars_dir + '/vaulted_vars')

print "Copying requirements file"
shutil.copy(checkout_dir + '/requirements.yml', root_dir + '/provisioning/requirements.yml')

def executeCommandRealtimeOutput(cmd):
	try:
	    print 'CMD:', ' '.join(cmd)
	    ps = subprocess.Popen((cmd), stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

	    lines_iterator = iter(ps.stdout.readline, b"")
	    for line in lines_iterator:
	        print(line) # yield line

		ps.returncode
	except subprocess.CalledProcessError as e:
	    print "error>", e.output, '<'
	    exit(1)


print "Installing third party ansible roles"
executeCommandRealtimeOutput(['ansible-galaxy', 'install', '-f', '-r', checkout_dir + '/requirements.yml', '-p',
                                 container_provisioning_roles_dir])

print "Executing lxc creation and provisioning"
os.chdir(root_dir + '/provisioning')
exit(executeCommandRealtimeOutput(['ansible-playbook', 'provision-container.yml']))
