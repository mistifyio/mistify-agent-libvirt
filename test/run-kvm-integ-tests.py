# Upload deb openvswitch to S3
# Clone infra jenkins slave repo
import git, os, shutil, subprocess

repo_name = 'infrastructure-jenkins-slave'
repo_remote = "git@github.com:mistifyio/" + repo_name
branch = 'master'
checkout_dir = repo_name
root_dir = os.path.dirname(os.path.realpath(__file__))

if os.path.isdir(checkout_dir):
    shutil.rmtree(checkout_dir)

repo = git.Repo.clone_from(repo_remote, checkout_dir, branch=branch)
container_provisioning_roles_dir = root_dir + '/provisioning/roles/'
infrastructure_roles = os.listdir(checkout_dir + '/roles/')

for role in infrastructure_roles:
    if os.path.isdir(container_provisioning_roles_dir + role):
    	shutil.rmtree(container_provisioning_roles_dir + role)
    print "Copying", role, "from", repo_name
    shutil.copytree(checkout_dir + '/roles/' + role, container_provisioning_roles_dir + role)

print "Copy vars files"
shutil.copy(checkout_dir + '/vars/vaulted_vars', root_dir + '/provisioning/group_vars/vaulted_vars')

print "Copy requirements file"
shutil.copy(checkout_dir + '/requirements.yml', root_dir + '/provisioning/requirements.yml')

print "Installing third party ansible roles"
executeCommand(['ansible-galaxy', 'install', '-f', '-r', checkout_dir + '/requirements.yml', '-p',
                                 container_provisioning_roles_dir])

print "Executing lxc creation and provisioning"
executeCommand(['ansible-playbook', 'provision-container.yml'])

print "Executing go kvm tests"
executeCommand(['ansible-playbook', 'execute-tests.yml'])

def executeCommand(cmd):
	try:
	    print 'CMD:', ' '.join(cmd)
	    subprocess.check_output(install_ansible_provision_cmd, stderr=subprocess.STDOUT)
	except subprocess.CalledProcessError as e:
	    print "error>", e.output, '<'
	    exit(1)