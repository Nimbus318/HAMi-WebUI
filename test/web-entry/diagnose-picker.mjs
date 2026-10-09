// Temporary fork-only instrumentation. No awaited work or console I/O in handlers.
import { readFileSync, writeFileSync } from 'node:fs'

function replaceOnce(source, before, after) {
  if (source.split(before).length !== 2) throw new Error(`Expected one match: ${before}`)
  return source.replace(before, after)
}

const pickerPath = 'packages/web/node_modules/tdesign-vue-next/es/date-picker/DateRangePicker.mjs'
let picker = readFileSync(pickerPath, 'utf8')
picker = replaceOnce(picker, '    var isSelected = ref(false);', `    var isSelected = ref(false);
    function tracePicker(event, extra) {
      if (typeof window !== "undefined" && window.__pickerTrace) {
        window.__pickerTrace.push(JSON.parse(JSON.stringify({
          event, now: performance.now(), activeIndex: activeIndex.value,
          input: inputValue.value, cache: cacheValue.value, time: time.value,
          applied: value.value, extra
        })));
      }
    }`)
picker = replaceOnce(picker, '    watch(popupVisible, function (visible) {\n      if (visible) {', '    watch(popupVisible, function (visible) {\n      tracePicker("popup-watch", { visible });\n      if (visible) {')
picker = replaceOnce(picker, '        if (!value.value.length) {', '        tracePicker("popup-reset-complete");\n        if (!value.value.length) {')
picker = replaceOnce(picker, '    function onCellMouseEnter(date) {', '    function onCellMouseEnter(date) {\n      tracePicker("cell-enter-before", { date });')
picker = replaceOnce(picker, '      inputValue.value = nextValue;\n    }\n    function onCellMouseLeave() {\n      isHoverCell.value = false;\n      inputValue.value = cacheValue.value;\n    }', '      inputValue.value = nextValue;\n      tracePicker("cell-enter-after", { date });\n    }\n    function onCellMouseLeave() {\n      tracePicker("cell-leave-before");\n      isHoverCell.value = false;\n      inputValue.value = cacheValue.value;\n      tracePicker("cell-leave-after");\n    }')
picker = replaceOnce(picker, '    function onTimePickerChange(val) {', '    function onTimePickerChange(val) {\n      tracePicker("time-change-before", { val });')
picker = replaceOnce(picker, '      cacheValue.value = formatDate(nextInputValue, {\n        format: formatRef.value.format\n      });\n    }', '      cacheValue.value = formatDate(nextInputValue, {\n        format: formatRef.value.format\n      });\n      tracePicker("time-change-after", { val });\n    }')
writeFileSync(pickerPath, picker)

const rangePath = 'packages/web/node_modules/tdesign-vue-next/es/date-picker/hooks/useRange.mjs'
let range = readFileSync(rangePath, 'utf8')
range = replaceOnce(range, '        inputValue.value = newVal;\n        if (!isValidDate(newVal, formatRef.value.format)) return;', `        inputValue.value = newVal;
        if (typeof window !== "undefined" && window.__pickerTrace) window.__pickerTrace.push(JSON.parse(JSON.stringify({ event: "range-input", now: performance.now(), input: inputValue.value, cache: cacheValue.value, time: time.value, newVal, valid: isValidDate(newVal, formatRef.value.format) })));
        if (!isValidDate(newVal, formatRef.value.format)) return;`)
range = replaceOnce(range, '        time.value = newTime;\n      },', '        time.value = newTime;\n        if (typeof window !== "undefined" && window.__pickerTrace) window.__pickerTrace.push(JSON.parse(JSON.stringify({ event: "range-input-complete", now: performance.now(), input: inputValue.value, cache: cacheValue.value, time: time.value })));\n      },')
writeFileSync(rangePath, range)
