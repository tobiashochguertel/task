#!/usr/bin/env zsh

# shellcheck shell=zsh
# shellcheck disable=all

echo "Quick test of task-dev..."
echo ""

CURR_DIR=$(pwd)
echo "Current directory: $CURR_DIR"
echo ""

cd "/Users/tobiashochgurtel/work-dev/my-projects/task/docs/transparent-mode/reference/vscode-demo-recorder" || (echo "Failed to cd to vscode-demo-recorder" && exit 1)

# Check if Taskfile.yml exists
if [ ! -f "Taskfile.yml" ]; then
	echo "Taskfile.yml does not exist"
	exit 1
else
	echo "Taskfile.yml exists"
fi

# Show task-dev binary location
echo "task-dev binary: $(which task-dev)"
echo ""

# If task-dev is a alias, unalias it
echo "Checking if task-dev is a alias..."
if [ ! -z "$(alias task-dev)" ]; then
	echo "task-dev is a alias"
	unalias task-dev
	echo "task-dev is no longer a alias"
	echo ""
else
	echo "task-dev is not a alias"
	echo ""
fi

echo "Checking task-dev version..."
task-dev --version
echo ""

# Set task-dev arguments
T_DEV_ARGS=(--transparent --show-whitespaces -v)

# Set tasks
T_TASKS=(
	debug
	prepublish
)

# Set additionally arguments
T_DEV_ADD_ARGS=(
	# "--transparent-renderer-table=custom"
	# "--transparent-renderer-table=lipgloss"

	'--transparent-renderer-table custom'
	'--transparent-renderer-table lipgloss'
)

# Set log directory
LOG_DIRNAME="logs"
# Set LOG_LOCATION
LOG_LOCATION="/Users/tobiashochgurtel/work-dev/my-projects/task/docs/transparent-mode/issues"

LOG_DIR="${LOG_LOCATION}/${LOG_DIRNAME}"

# Remove log directory if it exists
rm -rf "${LOG_DIR}" || echo "Failed to remove log directory or log directory does not exist: ${LOG_DIR}"

# Create log directory
mkdir -p "${LOG_DIR}" || (echo "Failed to create log directory: ${LOG_DIR}" && exit 1)

# Description:
#   Creates a log file name with the task name, additional arguments and timestamp
# Usage:
#   create_log_file_name <order number> <task name> <additional arguments>
# Example:
#   create_log_file_name 1 debug
#   create_log_file_name 1 debug "--transparent-renderer-table custom"
#
# Parameters:
#   $1: order number | required
#   $2: task name | required
#   $3: additional arguments | optional
# Return:
#   log file name
create_log_file_name() {
	local ORDER_NUMBER=$1
	local T_TASK=$2
	local T_DEV_ADD_ARG=$3

	# Check if Parameters are set
	if [ -z "$ORDER_NUMBER" ]; then
		echo "Error: Missing parameter 'ORDER_NUMBER'"
		exit 1
	fi
	if [ -z "$T_TASK" ]; then
		echo "Error: Missing parameter 'T_TASK'"
		exit 1
	fi

	# Remove spaces from task name
	T_TASK=$(echo "$T_TASK" | tr -d '=')

	# Remove spaces from additional arguments if set
	if [ ! -z "$T_DEV_ADD_ARG" ]; then
		T_DEV_ADD_ARG=$(echo "$T_DEV_ADD_ARG" | tr -d '=')
	else
		T_DEV_ADD_ARG="REMOVE-ME"
	fi

	# Create log file name
	# ${(l:10::0:)value}
	local LOG_FILENAME="${(l:2::0:)ORDER_NUMBER}-task-dev.${T_TASK}.${T_DEV_ADD_ARG}.$(date +%Y-%m-%d_%H-%M-%S).log"

	# Check if log_file includes 'REMOVE-ME', if so, remove it
	LOG_FILENAME=$(echo "$LOG_FILENAME" | sed 's/\.REMOVE-ME//g')

	# Check if log file name is valid
	if [ -z "$LOG_FILENAME" ]; then
		echo "Error: Invalid log file name: $LOG_FILENAME"
		exit 1
	fi

	# Return log file name
	echo "$LOG_FILENAME"
}

echo "Running tasks from 'Taskfile.yml' in '$CURR_DIR'..."
echo "Default Arguments for 'task-dev': ${T_DEV_ARGS[@]}"
echo ""

COUNT=1
# Run tasks
for T_TASK in $T_TASKS; do
	LOG_FILENAME=$(create_log_file_name $COUNT $T_TASK)
	echo "Running 'task-dev' with task '$T_TASK' and log file name '$LOG_FILENAME'..."
	task-dev "${T_DEV_ARGS[@]}" $T_TASK 2>&1 | tee "${LOG_DIR}/$LOG_FILENAME"
	LOG_FILENAME=""
	echo ""

	COUNT=$((COUNT + 1))

	# Iterate over additionally arguments
	for T_DEV_ADD_ARG in "${T_DEV_ADD_ARGS[@]}"; do
		LOG_FILENAME=$(create_log_file_name $COUNT $T_TASK $T_DEV_ADD_ARG)
		echo "Running 'task-dev' with task '$T_TASK' and additional argument '${T_DEV_ADD_ARG}' and log file name '$LOG_FILENAME'..."
		T_DEV_CMD="task-dev ${T_DEV_ARGS[@]} ${T_DEV_ADD_ARG} $T_TASK 2>&1 | tee \"${LOG_DIR}/$LOG_FILENAME\""
		echo "Running command: $T_DEV_CMD"
		eval $T_DEV_CMD
		LOG_FILENAME=""
		echo ""

		COUNT=$((COUNT + 1))
	done

	# COUNT=1
done

# Return to original directory
cd "$CURR_DIR" || (echo "Failed to cd back to $CURR_DIR" && exit 1)
