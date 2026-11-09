#!/usr/bin/env python3
"""Create a Windows PE RT_GROUP_ICON + RT_ICON .syso using clang's COFF assembler.
Runs on Linux, Mac or Windows with Python 3 and clang; no third-party Python deps.
The resulting icon is embedded into TimeStamps.exe so Explorer, taskbar, and
Windows Settings can display it without relying on a sidecar .ico file.
"""
import argparse
import os
import pathlib
import struct
import subprocess
import tempfile

def read_ico(path):
    data=path.read_bytes()
    reserved,kind,count=struct.unpack_from('<HHH',data,0)
    if reserved!=0 or kind!=1 or count<1: raise ValueError('Not an icon file')
    icons=[]
    group=bytearray(data[:6]); group=bytearray(struct.pack('<HHH',0,1,count))
    for i in range(count):
        w,h,colors,unused,planes,bits,sz,off=struct.unpack_from('<BBBBHHII',data,6+16*i)
        im=data[off:off+sz]
        if len(im)!=sz: raise ValueError('Truncated icon image')
        icons.append(im)
        group.extend(struct.pack('<BBBBHHIH',w,h,colors,unused,planes,bits,sz,i+1))
    return icons,bytes(group)

def resource_table(icons,group):
    # Resource directory root: type IDs 3=RT_ICON, 14=RT_GROUP_ICON.
    nodes=bytearray()
    offsets={}
    relocs=[]
    def newdir(name,count):
        off=len(nodes)
        nodes.extend(struct.pack('<IIHHHH',0,0,0,0,0,count))
        nodes.extend(b'\x00'*(8*count))
        offsets[name]=off
        return off
    root=newdir('root',2)
    type_icons=newdir('icon_type',len(icons))
    type_group=newdir('group_type',1)
    icon_lang=[]
    for i in range(len(icons)):
        icon_lang.append(newdir(f'icon_lang_{i+1}',1))
    group_lang=newdir('group_lang',1)
    def putentry(dir_off,n,index_or_id,where,isdir):
        struct.pack_into('<II',nodes,dir_off+16+n*8,index_or_id,where|(0x80000000 if isdir else 0))
    putentry(root,0,3,type_icons,True)
    putentry(root,1,14,type_group,True)
    for i,off in enumerate(icon_lang):putentry(type_icons,i,i+1,off,True)
    putentry(type_group,0,1,group_lang,True)
    entry_data_offsets=[]
    for i,off in enumerate(icon_lang):
        data_entry=len(nodes)
        putentry(off,0,1033,data_entry,False)
        nodes.extend(struct.pack('<IIII',0,len(icons[i]),0,0))
        entry_data_offsets.append((data_entry, f'icon_image_{i+1}'))
    group_entry=len(nodes)
    putentry(group_lang,0,1033,group_entry,False)
    nodes.extend(struct.pack('<IIII',0,len(group),0,0))
    entry_data_offsets.append((group_entry,'group_image'))
    return nodes,entry_data_offsets

def create_syso(ico,output):
    images,grp=read_ico(ico)
    data,relocs=resource_table(images,grp)
    blob_dir=pathlib.Path(tempfile.mkdtemp(prefix='ts_resource_'))
    try:
        segments=[]
        prev=0
        for off,sym in sorted(relocs):
            if off>prev:segments.append('.byte '+','.join(str(b) for b in data[prev:off]))
            segments.append(f'.long {sym}@IMGREL')
            prev=off+4
        if prev<len(data):segments.append('.byte '+','.join(str(b) for b in data[prev:]))
        asm=['.section .rsrc$01,"dr"','.p2align 2',*segments,'.section .rsrc$02,"dr"','.p2align 2']
        for i,image in enumerate(images):
            filename=blob_dir/f'icon_{i+1}.bin';filename.write_bytes(image)
            asm+= [f'icon_image_{i+1}:',f'.incbin "{filename}"','.p2align 2']
        fname=blob_dir/'group.bin';fname.write_bytes(grp)
        asm+=['group_image:',f'.incbin "{fname}"','.p2align 2']
        af=blob_dir/'resources.s';af.write_text('\n'.join(asm)+'\n')
        subprocess.run(['clang','-target','x86_64-windows-gnu','-c',str(af),'-o',str(output)],check=True)
    finally:
        import shutil
        shutil.rmtree(blob_dir)
    return len(images)

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('ico',type=pathlib.Path);p.add_argument('output',type=pathlib.Path)
    args=p.parse_args();print(f'Embedded {create_syso(args.ico,args.output)} icon sizes: {args.output}')
