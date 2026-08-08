// 找「誰改座標」—— 兩個問題,一支腳本。
//
//   tools/ida.sh idc WORRIORS.EXP.i64 ida_move_hunt.idc
//   → re_work/move_hunt.txt
//
// 問題 1(CanSail):`sub_2D2D0` 的「走一步」那一段沒有任何一行寫隊伍座標
//   (`docs/re/83` §5)。那麼**誰**在寫 `byte_3E0A6` / `byte_3E0A7`?
//
// 問題 2(sub_2E24):世界回合對每一槽只做攻擊(`sub_25F0`)與閘門(`sub_2D38`),
//   兩支都不改座標。那麼**誰**在寫物件槽的 +2 / +3(物件的 X / Y)?
//
// 為什麼一定要用 xref 而不是 grep `.asm`(`CLAUDE.md §4.5`):
//   物件座標的存取寫成 `byte ptr dword_3E46C+2[edx*8]` —— 32 個槽共用同一個
//   符號名,而槽號在暫存器裡。grep 找不到「第 7 槽的 Y」,但 xref 圖知道
//   每一筆參考落在哪個位址、是讀還是寫。
//
// headless 的 print/Message() 看不到 → 一律寫檔。
#include <idc.idc>

static kind_of(t) {
    // 以 IDA 標的型別為準,不要自己解析指令文字。
    if (t == 1) return "取址";
    if (t == 2) return "寫  ";
    if (t == 3) return "讀  ";
    return "其他";
}

// 列出一個位址的所有 dref。只印 want_writes 指定的那一類(0 = 全部)。
static dump_addr(out, label, ea, want_writes) {
    auto x, t, fn, n;
    if (ea == BADADDR) {
        fprintf(out, "== %s:找不到\n", label);
        return 0;
    }
    n = 0;
    for (x = get_first_dref_to(ea); x != BADADDR; x = get_next_dref_to(ea, x)) {
        t = XrefType();
        if (want_writes && t != 2) continue;
        fn = get_func_name(x);
        if (fn == "") fn = "(不在函式內)";
        fprintf(out, "  %s 0x%-8X %-14s %s\n", kind_of(t), x, fn, GetDisasm(x));
        n = n + 1;
    }
    return n;
}

static section(out, title) {
    fprintf(out, "\n################ %s ################\n", title);
}

static main() {
    auto out, i, ea, n, total;

    Wait();                                   // 一定要等自動分析跑完

    out = fopen("/work/move_hunt.txt", "w");
    if (out == 0) return;

    // ---- 問題 1:隊伍座標與帆的兩個變數 ----
    section(out, "問題 1:誰寫隊伍座標 / 帆的狀態");
    fprintf(out, "\n-- byte_3E0A6(隊伍 X)全部 dref --\n");
    n = dump_addr(out, "byte_3E0A6", get_name_ea_simple("byte_3E0A6"), 0);
    fprintf(out, "  共 %d 筆\n", n);

    fprintf(out, "\n-- byte_3E0A7(隊伍 Y)全部 dref --\n");
    n = dump_addr(out, "byte_3E0A7", get_name_ea_simple("byte_3E0A7"), 0);
    fprintf(out, "  共 %d 筆\n", n);

    fprintf(out, "\n-- byte_3E093(帆的步數計數器)全部 dref --\n");
    n = dump_addr(out, "byte_3E093", get_name_ea_simple("byte_3E093"), 0);
    fprintf(out, "  共 %d 筆\n", n);

    fprintf(out, "\n-- byte_3E167(帆向)全部 dref --\n");
    n = dump_addr(out, "byte_3E167", get_name_ea_simple("byte_3E167"), 0);
    fprintf(out, "  共 %d 筆\n", n);

    // ---- 問題 2:物件表 32 槽 × 8 B 的每一個位址 ----
    // 只印**寫入**,而且逐位址掃 —— 32 個槽共用符號名,所以要按位址分。
    section(out, "問題 2:誰寫物件槽的 X(+2)/ Y(+3)");
    ea = get_name_ea_simple("dword_3E46C");
    fprintf(out, "物件表基底 = 0x%X\n", ea);
    if (ea != BADADDR) {
        total = 0;
        for (i = 0; i < 32; i++) {
            n = 0;
            fprintf(out, "\n-- 槽 %d 的 +2(X)0x%X --\n", i, ea + i * 8 + 2);
            n = n + dump_addr(out, "X", ea + i * 8 + 2, 1);
            fprintf(out, "-- 槽 %d 的 +3(Y)0x%X --\n", i, ea + i * 8 + 3);
            n = n + dump_addr(out, "Y", ea + i * 8 + 3, 1);
            if (n == 0) fprintf(out, "  (無寫入)\n");
            total = total + n;
        }
        fprintf(out, "\n物件座標的寫入點合計 %d 筆\n", total);
    }

    // ---- 附帶:物件表整段的**所有**寫入(不分欄位),抓漏 ----
    section(out, "附帶:物件表 0x3E46C..+256 的所有寫入(不分欄位)");
    if (ea != BADADDR) {
        total = 0;
        for (i = 0; i < 256; i++) {
            n = dump_addr(out, "", ea + i, 1);
            if (n > 0) fprintf(out, "   ↑ 偏移 +%d\n", i);
            total = total + n;
        }
        fprintf(out, "\n合計 %d 筆寫入\n", total);
    }

    fclose(out);
}
