import re, os
names = {}
for root in ('internal/game','internal/u5data'):
    for fn in sorted(os.listdir(root)):
        if not fn.endswith('.go') or fn.endswith('_test.go'):
            continue
        p = os.path.join(root, fn)
        for m in re.finditer(r'^func \(\w+ \*?\w+\) (\w+)\(', open(p,encoding='utf-8').read(), re.M):
            names.setdefault(m.group(1), []).append(p)
body, tests = '', ''
for root in ('internal','cmd','tools'):
    for dirpath, _, files in os.walk(root):
        for fn in files:
            if not fn.endswith('.go'):
                continue
            t = open(os.path.join(dirpath,fn),encoding='utf-8').read()
            if fn.endswith('_test.go'):
                tests += t
            else:
                body += t
dead = []
for n, where in sorted(names.items()):
    if len(re.findall(r'\.'+re.escape(n)+r'\(', body)) == 0:
        dead.append((n, where[0], len(re.findall(r'\.'+re.escape(n)+r'\(', tests))))
print("=== 沒有任何非測試呼叫者的方法 ===")
for n, w, t in dead:
    print("  %-34s %-40s 測試引用 %d" % (n, w, t))
print(len(dead), "個")
