import {useState} from 'react'
import './App.css'
import {
    Alert,
    BottomNavigation,
    BottomNavigationAction,
    Box,
    Breadcrumbs,
    Button,
    Link,
    List,
    ListItem,
    ListItemAvatar,
    ListItemButton,
    ListItemText,
    Snackbar,
    TextField,
    Typography
} from "@mui/material";
import Grid2 from "@mui/material/Unstable_Grid2";
import FolderIcon from '@mui/icons-material/Folder';


function Filelist(props: { fileList: Array<{ type: number, filename: string }>, onClick: (filename: string) => void }) {
    return <List>
        {props.fileList.map((value =>
            <ListItemButton onClick={(e) => {
                props.onClick(value.filename)
            }}>
                <ListItem>
                    <ListItemAvatar>
                        {value.type == 1 && <FolderIcon/>}
                    </ListItemAvatar><ListItemText>{value.filename}</ListItemText></ListItem>
            </ListItemButton>))}
    </List>
}

function App() {
    const [page, setPage] = useState(0)
    const [clipBoard, setClipBoard] = useState("")
    const [fileList, setFileList] = useState([{type: 1, filename: "2333d"}, {type: 2, filename: "2333"}])
    const [nowPath, setNowPath] = useState(["Upload", "Pic", "Wallpaper"])
    const [open, setOpen] = useState(false)
    const [message, setMessage] = useState("")

    let ws = new WebSocket("ws://" + document.location.host + "/ws");
    let wsCloseFunc = (e: any) => {
        console.log(e)
        ws = new WebSocket("ws://" + document.location.host + "/ws")
        ws.onmessage = (e) => {
            var data = JSON.parse(e.data)
            console.log(data)
            setClipBoard(data.message)
        }
        ws.onclose = wsCloseFunc
    };
    ws.onmessage = (e) => {
        var data = JSON.parse(e.data)
        console.log(data)
        setClipBoard(data.message)
    }
    ws.onclose = wsCloseFunc

    const handleClose = (e: any) => {
        setOpen(false)
        setMessage("")
    }

    return (
        <Box sx={{}}>
            <Grid2 container spacing={2}>
                <Grid2 xs={12}>
                    {/*page 1*/}
                    {page == 0 && <TextField label={"ClipBoard In PC"}
                                             variant="filled"
                                             multiline
                                             rows={30}
                                             sx={{height: "90vh", width: "99%"}}
                                             value={clipBoard}
                                             onChange={(event) => {
                                                 setClipBoard(event.target.value);
                                                 fetch("/set", {
                                                     method: "POST",
                                                     headers: {
                                                         'Content-Type': 'application/json'
                                                     },
                                                     body: JSON.stringify({
                                                         message: event.target.value
                                                     })
                                                 }).then((res) => {
                                                         if (res.status === 200) {
                                                             setMessage("Paste Success")
                                                             setOpen(true)
                                                         } else {
                                                             setMessage("Paste Failed")
                                                             setOpen(true)
                                                         }
                                                     }
                                                 )
                                             }}
                                             onFocus={(event) => {
                                                 event.target.select();
                                             }}/>}
                    {/*page 2*/}

                    {page == 3 && <>
                        <Breadcrumbs>
                            {nowPath.map((value, index) => {
                                    if (index === nowPath.length - 1) {
                                        return <Typography color="text.primary">{value}</Typography>
                                    } else {
                                        return <Link underline="hover" color="inherit"
                                                     onClick={() => setNowPath(nowPath.slice(0, index + 1))}>{value}</Link>
                                    }
                                }
                            )}

                        </Breadcrumbs>
                        <Filelist fileList={fileList} onClick={(filename) => setNowPath([...nowPath, filename])}/>
                    </>
                    }
                    {page == 1 && <iframe src={"/download"} height={"90vw"} width={"100%"}/>

                    }

                    {/*    Page 3*/}
                    {page == 2 && <Button color={"primary"} variant="contained" component="label">
                        Upload
                        <input hidden type="file" onChange={(event => {
                            let file = event!.target!.files![0];
                            if(!file) return;
                            var fd = new FormData()
                            fd.append("upload", file)
                            fetch("/upload", {
                                method: "POST",
                                body: fd
                            }).then((res) => {
                                if (res.status === 200) {
                                    setMessage("Upload Success")
                                    setOpen(true)
                                } else {
                                    setMessage("Upload Failed")
                                    setOpen(true)
                                }

                            })

                        })}/>
                    </Button>}
                </Grid2>
            </Grid2><BottomNavigation
            showLabels
            value={page}
            onChange={(event, value) => {
                setPage(value);
            }}
            sx={{position: 'fixed', bottom: 0, left: 0, right: 0}}
        >
            <BottomNavigationAction label={"ClipBoard"}/>
            <BottomNavigationAction label={"Download File"}/>
            <BottomNavigationAction label={"Upload File"}/>
        </BottomNavigation>
            <Snackbar open={open} autoHideDuration={6000} onClose={handleClose}>
                <Alert onClose={handleClose} severity="success" sx={{width: '100%'}}>
                    {message}
                </Alert>
            </Snackbar>
        </Box>


    )
}


export default App
