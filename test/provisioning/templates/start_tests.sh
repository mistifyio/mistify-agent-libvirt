export WORKSPACE=$HOME/gotest_workspace
export REPO={{mistify_agent_repo_name}}
mkdir $WORKSPACE
cd $WORKSPACE
git clone {{mistify_agent_repo_url}} $WORKSPACE/src/{{mistify_agent_repo_name}}
cd $WORKSPACE/{{mistify_agent_repo_name}}
python /tmp/go_test_executor.py
