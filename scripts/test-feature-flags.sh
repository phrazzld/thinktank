#!/bin/bash
set -euo pipefail

# Test script for quality gate feature flags
# This script tests that the feature flag system works correctly

CONFIG_FILE=".github/quality-gates-config.yml"
ACTION_PATH=".github/actions/read-quality-gate-config"

echo "🧪 Testing Quality Gate Feature Flags"
echo "======================================="

# Function to test config reading with a given config
test_config_reading() {
    local test_name="$1"
    local config_content="$2"
    local expected_lint_enabled="$3"
    local expected_secret_scan_enabled="$4"

    echo "📝 Test: $test_name"

    # Create temporary config file
    local temp_config="/tmp/test-quality-gates-config.yml"
    echo "$config_content" > "$temp_config"

    # Create a temporary test action that uses our config parser
    local test_script="/tmp/test-config-reader.sh"
    cat > "$test_script" << 'EOF'
#!/bin/bash
set -euo pipefail

CONFIG_FILE="$1"

# Function to extract YAML values safely
get_yaml_value() {
  local key="$1"
  local file="$2"
  local default="${3:-false}"

  # Use yq if available, otherwise fallback to grep/sed
  if command -v yq >/dev/null 2>&1; then
    value=$(yq eval "$key" "$file" 2>/dev/null || echo "$default")
  else
    # Simple grep-based parser for basic YAML
    value=$(grep -A 10 "$key" "$file" | grep -E "^\s*(enabled|required):" | head -1 | sed 's/.*: *//' | tr -d ' ' || echo "$default")
  fi

  # Normalize boolean values
  case "$value" in
    true|True|TRUE|yes|Yes|YES|1) echo "true" ;;
    false|False|FALSE|no|No|NO|0) echo "false" ;;
    *) echo "$default" ;;
  esac
}

# Install yq for YAML parsing if not available
if ! command -v yq >/dev/null 2>&1; then
  echo "Installing yq for YAML parsing..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    if command -v brew >/dev/null 2>&1; then
      brew install yq >/dev/null 2>&1 || true
    else
      curl -L https://github.com/mikefarah/yq/releases/latest/download/yq_darwin_amd64 -o /usr/local/bin/yq
      chmod +x /usr/local/bin/yq
    fi
  else
    # Linux
    sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
    sudo chmod +x /usr/local/bin/yq
  fi
fi

# Test reading values
lint_enabled=$(get_yaml_value '.ci_gates.lint.enabled' "$CONFIG_FILE" 'true')
secret_scan_enabled=$(get_yaml_value '.security_gates.secret_scan.enabled' "$CONFIG_FILE" 'true')

echo "lint_enabled=$lint_enabled"
echo "secret_scan_enabled=$secret_scan_enabled"
EOF

    chmod +x "$test_script"

    # Run the test and capture output
    output=$("$test_script" "$temp_config")

    # Extract values from output
    lint_enabled=$(echo "$output" | grep "lint_enabled=" | cut -d'=' -f2)
    secret_scan_enabled=$(echo "$output" | grep "secret_scan_enabled=" | cut -d'=' -f2)

    # Verify expected values
    if [[ "$lint_enabled" == "$expected_lint_enabled" ]] && [[ "$secret_scan_enabled" == "$expected_secret_scan_enabled" ]]; then
        echo "  ✅ PASS: lint_enabled=$lint_enabled, secret_scan_enabled=$secret_scan_enabled"
    else
        echo "  ❌ FAIL: Expected lint_enabled=$expected_lint_enabled, secret_scan_enabled=$expected_secret_scan_enabled"
        echo "          Got lint_enabled=$lint_enabled, secret_scan_enabled=$secret_scan_enabled"
        return 1
    fi

    # Cleanup
    rm -f "$temp_config" "$test_script"
}

# Test 1: Default configuration (all enabled)
test_config_reading "Default Configuration" \
'version: "1.0"
ci_gates:
  lint:
    enabled: true
    required: true
security_gates:
  secret_scan:
    enabled: true
    required: true' \
"true" "true"

# Test 2: Lint disabled, secret scan enabled
test_config_reading "Lint Disabled" \
'version: "1.0"
ci_gates:
  lint:
    enabled: false
    required: true
security_gates:
  secret_scan:
    enabled: true
    required: true' \
"false" "true"

# Test 3: Both disabled
test_config_reading "Both Disabled" \
'version: "1.0"
ci_gates:
  lint:
    enabled: false
    required: false
security_gates:
  secret_scan:
    enabled: false
    required: false' \
"false" "false"

# Test 4: Test that the actual config file is valid
echo "📝 Test: Validate actual config file"
if [[ -f "$CONFIG_FILE" ]]; then
    # Try to read the actual config file
    if command -v yq >/dev/null 2>&1; then
        yq eval '.ci_gates.lint.enabled' "$CONFIG_FILE" >/dev/null
        echo "  ✅ PASS: Actual config file is valid YAML"
    else
        echo "  ⚠️  SKIP: yq not available, cannot validate YAML syntax"
    fi
else
    echo "  ❌ FAIL: Config file $CONFIG_FILE not found"
    exit 1
fi

# Test 5: Test that the action file exists and is properly structured
echo "📝 Test: Validate action file structure"
if [[ -f "$ACTION_PATH/action.yml" ]]; then
    if grep -q "name: 'Read Quality Gate Configuration'" "$ACTION_PATH/action.yml"; then
        echo "  ✅ PASS: Action file exists and has correct structure"
    else
        echo "  ❌ FAIL: Action file exists but missing expected content"
        exit 1
    fi
else
    echo "  ❌ FAIL: Action file $ACTION_PATH/action.yml not found"
    exit 1
fi

echo ""
echo "🎉 All feature flag tests passed!"
echo ""
echo "💡 To test in GitHub Actions:"
echo "   1. Create a PR that modifies .github/quality-gates-config.yml"
echo "   2. Disable a gate (e.g., set ci_gates.lint.enabled: false)"
echo "   3. Verify that the corresponding job is skipped in the workflow"
echo ""
echo "🔧 To disable a quality gate:"
echo "   - Edit $CONFIG_FILE"
echo "   - Set the desired gate's 'enabled' field to false"
echo "   - Commit and push to see the effect in CI"
