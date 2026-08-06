// 查一個具名符號的所有交叉參考(讀 / 寫 / 取址)+ 所屬函式。
//
//   tools/ida.sh idc WORRIORS.EXP.i64 ida_xref.idc off_41BA0
//
// 為什麼要用這支而不是 grep 反編譯輸出或 .asm:
//   - 反編譯輸出裡,間接存取(基址 + 索引、指標算術)看不到符號名 —— grep 會回零命中,
//     而零命中和「真的沒人用」長得一模一樣。
//   - .asm 是攤平的文字,沒有交叉參考圖。
//   - IDA 建庫時就把型別標好了,直接問 XrefType() 最準;自己解析指令文字會判錯
//     (助憶碼後面補多個空格、push 的第 0 個運算元是來源不是目的)。
//
// headless 的 print/Message() 看不到,所以結果一律寫檔。
#include <idc.idc>

static kind_of(t) {
    // dr_O=1 取址 / dr_W=2 寫 / dr_R=3 讀(以 IDA 標的型別為準,不要自己猜指令語意)
    if (t == 1) return "取址";
    if (t == 2) return "寫  ";
    return "讀  ";
}

static main() {
    auto name, ea, x, out, n, fn, t;

    Wait();                                  // 一定要等自動分析跑完

    name = "";
    if (ARGV.count > 1) name = ARGV[1];
    out = fopen("/work/xref_out.txt", "w");
    if (out == 0) return;

    if (name == "") {
        fprintf(out, "用法:idat -A \"-S/work/tools/ida_xref.idc <符號名>\" <db>.i64\n");
        fclose(out);
        return;
    }

    ea = get_name_ea_simple(name);
    fprintf(out, "符號 %s → %s\n", name, ea == BADADDR ? "找不到" : sprintf("0x%X", ea));
    if (ea == BADADDR) {
        fclose(out);
        return;
    }

    n = 0;
    fprintf(out, "\n-- 資料參考(dref)--\n");
    for (x = get_first_dref_to(ea); x != BADADDR; x = get_next_dref_to(ea, x)) {
        t = XrefType();
        fn = get_func_name(x);
        if (fn == "") fn = "(不在函式內)";
        fprintf(out, "%s 0x%-8X %-22s %s\n", kind_of(t), x, fn, GetDisasm(x));
        n = n + 1;
    }
    fprintf(out, "共 %d 筆 dref\n", n);

    n = 0;
    fprintf(out, "\n-- 程式碼參考(cref,呼叫/跳轉到此)--\n");
    for (x = get_first_cref_to(ea); x != BADADDR; x = get_next_cref_to(ea, x)) {
        fn = get_func_name(x);
        if (fn == "") fn = "(不在函式內)";
        fprintf(out, "     0x%-8X %-22s %s\n", x, fn, GetDisasm(x));
        n = n + 1;
    }
    fprintf(out, "共 %d 筆 cref\n", n);

    // 提醒:xref 只涵蓋 IDA 認得出來的參考。把位址當純數字算的程式碼
    // (mov eax, <位址> 之類)不在圖上;若結果異常少,要另掃立即數。
    fclose(out);
}
