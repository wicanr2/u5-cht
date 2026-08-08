#include <idc.idc>
// CD-ROM 音軌播放的呼叫鏈。位址 = 檔案位移 − 0x200(已驗證)。
static show(f, nm, ea) {
  auto fn, x, n, pf;
  fprintf(f, "\n=== %s  banner@0x%05X\n", nm, ea);
  fn = get_func_attr(ea, FUNCATTR_START);
  if (fn == BADADDR || fn == -1) {
    pf = ea;
    while (pf > 0x1000 && (get_func_attr(pf, FUNCATTR_START) == BADADDR ||
                           get_func_attr(pf, FUNCATTR_START) == -1))
      pf = pf - 1;
    fn = get_func_attr(pf, FUNCATTR_START);
    fprintf(f, "    banner 在函式外;往前找到 %s(0x%05X)\n", GetFunctionName(fn), fn);
  } else {
    fprintf(f, "    banner 在 %s(0x%05X) 之內\n", GetFunctionName(fn), fn);
  }
  n = 0;
  for (x = RfirstB(fn); x != BADADDR; x = RnextB(fn, x)) {
    fprintf(f, "    ← 呼叫者 %s @ 0x%05X\n", GetFunctionName(x), x);
    n = n + 1;
    if (n > 30) break;
  }
  if (n == 0) fprintf(f, "    (該函式沒有直接呼叫者)\n");
}
static main() {
  auto f;
  Wait();
  f = fopen("/work/cdda_hunt.txt", "w");
  show(f, "cdr_status",  0x3348C);
  show(f, "cdr_mstop",   0x33549);
  show(f, "cdr_mtplay",  0x336FF);
  show(f, "cdr_rmtplay", 0x3376B);
  show(f, "cdr_mphase",  0x337D3);
  show(f, "cdr_cdinfo",  0x3385A);
  fclose(f);
}
