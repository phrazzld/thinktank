#!/bin/bash
# Script to analyze Claude warnings log

LOG_FILE=".claude-warnings.log"

if [ ! -f "$LOG_FILE" ]; then
  echo "No Claude warnings log found. Run a commit with warnings first."
  exit 0
fi

show_help() {
  echo "Usage: $(basename "$0") [OPTIONS]"
  echo "Analyze the Claude warnings log"
  echo ""
  echo "Options:"
  echo "  -h, --help      Show this help message"
  echo "  -l, --list      List all entries with their status and commit hash"
  echo "  -c HASH         Show details for a specific commit hash"
  echo "  -s STATUS       Filter by status (WARN, FAIL)"
  echo "  -b BRANCH       Filter by branch name"
  echo "  --summary       Show a summary of warnings/failures by branch"
  echo ""
  echo "Examples:"
  echo "  $(basename "$0") --list                # List all entries"
  echo "  $(basename "$0") -c abc123             # Show details for commit abc123"
  echo "  $(basename "$0") -s WARN               # Show all warnings"
  echo "  $(basename "$0") -b feature/branch     # Show entries for a branch"
  echo "  $(basename "$0") --summary             # Show summary statistics"
}

list_entries() {
  echo "Claude warnings/failures log entries:"
  echo "----------------------------------------"

  # Print header with field names
  echo "STATUS   COMMIT          DATE                     CORRELATION_ID                           MESSAGE"

  grep -A1 -B1 "STATUS: " "$LOG_FILE" | grep -v "^--$" | \
  awk 'BEGIN {RS="--"; FS="\n"}
       NF>2 {
         for(i=1; i<=NF; i++) {
           if ($i ~ /STATUS:/) status=$i;
           if ($i ~ /COMMIT:/) commit=$i;
           if ($i ~ /BRANCH:/) branch=$i;
           if ($i ~ /COMMIT_MSG:/) msg=$i;
           if ($i ~ /DATE:/) date=$i;
           if ($i ~ /CORRELATION_ID:/) corr_id=$i;
         }
         sub("STATUS: ", "", status);
         sub("COMMIT: ", "", commit);
         sub("BRANCH: ", "", branch);
         sub("COMMIT_MSG: ", "", msg);
         sub("DATE: ", "", date);
         sub("CORRELATION_ID: ", "", corr_id);
         printf "%-8s %-15s %-25s %-36s %s\n", status, commit, date, corr_id, msg;
       }'
}

show_commit_details() {
  local commit_hash=$1
  echo "Details for commit: $commit_hash"
  echo "----------------------------------------"
  sed -n "/COMMIT: $commit_hash/,/^$/p" "$LOG_FILE"
}

filter_by_status() {
  local status=$1
  echo "Entries with status: $status"
  echo "----------------------------------------"

  awk -v status="$status" 'BEGIN {RS="----------------------------------------"; FS="\n"; ORS="\n\n"}
       $0 ~ "STATUS: "status {
         print "----------------------------------------" $0 "----------------------------------------";
       }' "$LOG_FILE"
}

filter_by_branch() {
  local branch=$1
  echo "Entries for branch: $branch"
  echo "----------------------------------------"

  awk -v branch="$branch" 'BEGIN {RS="----------------------------------------"; FS="\n"; ORS="\n\n"}
       $0 ~ "BRANCH: "branch {
         print "----------------------------------------" $0 "----------------------------------------";
       }' "$LOG_FILE"
}

show_summary() {
  echo "Summary of Claude warnings and failures:"
  echo "----------------------------------------"

  # Count total warnings and failures
  local warn_count=$(grep -c "STATUS: WARN" "$LOG_FILE")
  local fail_count=$(grep -c "STATUS: FAIL" "$LOG_FILE")

  echo "Total warnings: $warn_count"
  echo "Total failures: $fail_count"
  echo ""

  # Count by branch
  echo "Warnings/failures by branch:"
  grep "BRANCH: " "$LOG_FILE" | sort | uniq -c | sort -nr | \
  while read -r count branch; do
    branch=${branch#BRANCH: }
    echo "  $branch: $count"
  done

  echo ""

  # Most common warning/failure types (simplified, may need improvement)
  echo "Most common issue types:"
  grep -A2 "STATUS: " "$LOG_FILE" | grep -v "STATUS: " | grep -v "COMMIT: " | grep -v "^--$" | \
  grep -v "DATE: " | grep -v "BRANCH: " | grep -v "COMMIT_MSG: " | sort | uniq -c | sort -nr | head -5 | \
  while read -r count issue; do
    echo "  $count: $issue"
  done
}

# Main script execution
if [ $# -eq 0 ]; then
  show_help
  exit 0
fi

while [ $# -gt 0 ]; do
  case "$1" in
    -h|--help)
      show_help
      exit 0
      ;;
    -l|--list)
      list_entries
      exit 0
      ;;
    -c)
      shift
      if [ -n "$1" ]; then
        show_commit_details "$1"
      else
        echo "Error: -c requires a commit hash"
        exit 1
      fi
      exit 0
      ;;
    -s)
      shift
      if [ -n "$1" ]; then
        filter_by_status "$1"
      else
        echo "Error: -s requires a status (WARN or FAIL)"
        exit 1
      fi
      exit 0
      ;;
    -b)
      shift
      if [ -n "$1" ]; then
        filter_by_branch "$1"
      else
        echo "Error: -b requires a branch name"
        exit 1
      fi
      exit 0
      ;;
    --summary)
      show_summary
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      show_help
      exit 1
      ;;
  esac
  shift
done

exit 0
