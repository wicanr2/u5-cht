#include <idc.idc>
// 把 sub_27F0 用的 11×11 遮罩表 byte_4FD98 逐位元組 dump 出來。
// ⚠ 不從 .asm 的 `dd`/`align` 手抄 —— 風向表就是那樣抄錯三處的。
static main() {
  auto f, i, a, v;
  Wait();
  f = fopen("/work/mask_dump.txt", "w");
  a = 0x4FD98;
  fprintf(f, "byte_4FD98 起 72 byte(索引 = 11*dy + dx,dx/dy 各 0..5 ⇒ 上限 60)\n");
  for (i = 0; i < 72; i++) {
    v = Byte(a + i);
    fprintf(f, "%3d 0x%05X %02X\n", i, a + i, v);
  }
  // 順帶確認清場與移動閘門用到的兩個切換位元的初值
  fprintf(f, "\nbyte_4FDD5=%02X byte_4FDD7=%02X\n", Byte(0x4FDD5), Byte(0x4FDD7));
  // 以及 sub_27F0 的呼叫者確認(誰在用 0xFC 那一支)
  fprintf(f, "\nsub_27F0 的呼叫者:\n");
  a = get_name_ea(BADADDR, "sub_27F0");
  for (v = RfirstB(a); v != BADADDR; v = RnextB(a, v))
    fprintf(f, "  0x%05X %s (%s)\n", v, GetFunctionName(v), XrefTypeName(XrefType()));
  fclose(f);
}
