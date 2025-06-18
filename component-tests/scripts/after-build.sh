      if [ -f "./build/reports/cucumber.json" ]; then
        echo "Test Results Summary:"
        if grep -q '"status":"failed"' ./build/reports/cucumber.json; then
          echo "❌ Some tests FAILED"
          failed_count=$(grep -o '"status":"failed"' ./build/reports/cucumber.json | wc -l)
          echo "   Failed tests count: $failed_count"
        else
          echo "🎉 All tests PASSED"
        fi
        passed_count=$(grep -o '"status":"passed"' ./build/reports/cucumber.json | wc -l)
        echo "   Passed tests count: $passed_count"
        total_count=$(($(grep -o '"status":"' ./build/reports/cucumber.json | wc -l)))
        echo "   Total tests count: $total_count"
      else
        echo "⚠️ No test results found in ./build/reports/cucumber.json"
      fi